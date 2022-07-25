package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "9988"
	SERVER_TYPE = "tcp"
)

type Client struct {
	ID   string
	Conn net.Conn
	Pool *Pool
}

type Message struct {
	Type int    `json:"type"`
	Body string `json:"body"`
}

type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan Message
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan Message),
	}
}

func (pool *Pool) Start() {
	for {
		select {
		case clientR := <-pool.Register:
			pool.Clients[clientR] = true
			fmt.Println("Size of connection pool: ", len(pool.Clients))
			for client, _ := range pool.Clients {
				if client.Conn == clientR.Conn {
					continue
				}
				encoder := json.NewEncoder(client.Conn)
				if err := encoder.Encode(Message{Type: 1, Body: "New User Joined..."}); err != nil {
					fmt.Println(err)
					return
				}
			}
			break
		case clientU := <-pool.Unregister:
			delete(pool.Clients, clientU)
			fmt.Println("Size of connection pool: ", len(pool.Clients))
			for client, _ := range pool.Clients {
				if client.Conn == clientU.Conn {
					continue
				}
				encoder := json.NewEncoder(client.Conn)
				if err := encoder.Encode(Message{Type: 1, Body: "User disconected"}); err != nil {
					fmt.Println(err)
					return
				}
				//client.Conn.WriteJSON(Message{Type: 1, Body: "User disconected"})
			}
			break
		case message := <-pool.Broadcast:
			fmt.Println("Sending message to all clients", message)
			for client, _ := range pool.Clients {
				encoder := json.NewEncoder(client.Conn)
				if err := encoder.Encode(message); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
}

func (c *Client) Read() {
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	for {
		var msg Message
		if err := json.NewDecoder(c.Conn).Decode(&msg); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Received: ", msg.Body)
		message := msg
		c.Pool.Broadcast <- message
	}
}

func main() {
	fmt.Println("Server Running..")
	pool := NewPool()
	go pool.Start()
	server, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	if err != nil {
		fmt.Println("Error listening: " + err.Error())
		os.Exit(1)
	}
	defer server.Close()
	for {
		fmt.Println("Listening on " + SERVER_HOST + ":" + SERVER_PORT)
		fmt.Println("Waiting for client...")
		conn, err := server.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("Client " + conn.RemoteAddr().String() + "connected")
		fmt.Println("Start typing: ")
		go processClient(pool, conn)
	}
}

func processClient(pool *Pool, conn net.Conn) {

	client := Client{
		Conn: conn,
		Pool: pool,
	}
	pool.Register <- &client
	client.Read()
}
