package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type Command struct {
	numberElements int
	elements       []string
}

func newSimpleString(str string) string {
	return fmt.Sprintf("+%v\r\n", str)
}

func newBulkString(str string) string {
	return fmt.Sprintf("$%d\r\n%v\r\n", len(str), str)
}

func pong() []byte {
	return []byte("+PONG\r\n")
}

func echo(str string) []byte {
	return []byte(newBulkString(str))
}

func parser(str string) ([]byte, error) {
	command := strings.Split(str, "\r\n")
	fmt.Printf("command: %v\n", command)

	if command[0][0] != '*' {
		fmt.Println("Not a valid command ig")
	}

	formattedCommand := strings.ToLower(command[2])

	switch formattedCommand {
	case "ping":
		return pong(), nil
	case "echo":
		return echo(command[4]), nil
	}

	return nil, errors.New("Command unsupported")
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Thread created")

	for {
		buf := make([]byte, 1024)
		count, err := conn.Read(buf)
		if errors.Is(err, io.EOF) {
			fmt.Println("Client closed the connections:", conn.RemoteAddr())
			break
		} else if err != nil {
			fmt.Println("Error reading conn: ", err.Error())
			os.Exit(1)
		}

		if count == 0 {
			return
		}

		data := string(buf[:count])

		response, err := parser(data)
		if err != nil {
			fmt.Println(err)
			break
		}
		conn.Write(response)
	}
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		fmt.Println("Connection received, starting thread")
		go handleConnection(conn)
	}
}
