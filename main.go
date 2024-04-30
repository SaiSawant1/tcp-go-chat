package main

import (
	"log"
	"net"
	"time"
)

const SAFEMODE = false

func SafeMode(address string) string {
	if SAFEMODE == true {
		return "[IP HIDDEN]"
	}
	return address
}

type MessageType int

const (
	ClientConnected MessageType = iota + 1
	ClientNewMessge
	ClientDisconnected
)

type Message struct {
	Type    MessageType
	Message string
	Conn    net.Conn
}

func newMessage(messageType MessageType, message string, conn net.Conn) Message {
	return Message{Type: messageType, Message: message, Conn: conn}
}

type Client struct {
	Conn        net.Conn
	IsBanned    bool
	LastMessage time.Time
	BannedAt    time.Time
}

func (c *Client) BanClient() {
	c.IsBanned = true
}

func client(conn net.Conn, messages chan<- Message) {
	buffer := make([]byte, 512)
	message := newMessage(MessageType(1), "", conn)
	messages <- message
	for {
		_, err := conn.Read(buffer)
		if err != nil {
			messages <- newMessage(ClientDisconnected, "", conn)
			return
		}
		message = newMessage(MessageType(2), string(buffer), conn)
		messages <- message
	}
}

func server(messages <-chan Message) {
	clients := map[string]Client{}
	for {
		msg := <-messages
		switch msg.Type {
		// client is connected
		case ClientConnected:
			clients[msg.Conn.RemoteAddr().String()] = Client{Conn: msg.Conn, IsBanned: false, LastMessage: time.Now()}
			msg.Conn.Write([]byte("Welcome new User\n"))

			// client sends new message
		case ClientNewMessge:
			for _, client := range clients {
				if client.Conn.RemoteAddr() != msg.Conn.RemoteAddr() {
					_, err := client.Conn.Write([]byte(msg.Message))
					if err != nil {
						log.Println(err)
					}
				}
			}
			// client is disconnected
		case ClientDisconnected:
			delete(clients, msg.Conn.RemoteAddr().String())
			msg.Conn.Close()
		}
	}
}

func main() {
	const PORT = ":8080"
	ln, err := net.Listen("tcp", PORT)
	if err != nil {
		log.Fatalf("[ERROR] could not listen on port %s: %s\n", PORT, err)
	}
	log.Printf("Listent to TCP connections on PORT:%s\n", PORT)

	messages := make(chan Message)
	go server(messages)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[ERROR] could not connect to %s:%s\n", SafeMode(ln.Addr().String()), err)
		}
		log.Printf("Accepted Connection from %s", SafeMode(conn.RemoteAddr().String()))
		go client(conn, messages)
	}
}
