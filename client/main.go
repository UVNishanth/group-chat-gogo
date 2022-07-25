package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

type Message struct {
	Type int    `json:"type"`
	Body string `json:"body"`
}

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "9988"
	SERVER_TYPE = "tcp"
)

func main() {
	conn, err := net.Dial(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	if err != nil {
		panic(err)
	}
	fmt.Println("Connection established. Now talk....")
	go receive(conn)
	for {
		// send some data
		reader := bufio.NewReader(os.Stdin)
		//var msg string
		msg, _ := reader.ReadString('\n')
		msg = strings.TrimSuffix(msg, "\n")
		encoder := json.NewEncoder(conn)
		if err := encoder.Encode(Message{Type: 1, Body: msg}); err != nil {
			fmt.Println(err)
			return
		}
	}
}

func receive(conn net.Conn) {
	defer conn.Close()
	for {
		// buffer := make([]byte, 1024)
		// mLen, err := conn.Read(buffer)
		// if err != nil {
		// 	fmt.Println("Error reading: ", err.Error())
		// }
		// rcv := string(buffer[:mLen])
		// fmt.Println("Received: ", rcv)
		// if strings.Contains(rcv, "Bye") {
		// 	//conn.Write([]byte("Accha Bye Bye"))
		// 	break
		// }
		var msg Message
		if err := json.NewDecoder(conn).Decode(&msg); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Received: ", msg.Body)
	}
}
