package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	for {
		bytes := []byte{}
		_, err := conn.Read(bytes)
		if err != nil {
			fmt.Println("Error reading conn: ", err.Error())
			os.Exit(1)
		}

		fmt.Println("Received: ", bytes)
		response := "+PONG\r\n"
		conn.Write([]byte(response))
	}
}
