package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	NULL_BULK_STRING = "$-1\r\n"
	OK               = "OK"
)

func NewSimpleString(str string) string {
	return fmt.Sprintf("+%v\r\n", str)
}

func NewBulkString(str string) string {
	if str == "" {
		return NULL_BULK_STRING
	}
	return fmt.Sprintf("$%d\r\n%v\r\n", len(str), str)
}

func pong() []byte {
	return []byte("+PONG\r\n")
}

func echo(str string) []byte {
	return []byte(NewBulkString(str))
}

func parser(str string, store Store) ([]byte, error) {
	trimmed := strings.TrimSpace(str)
	command := strings.Split(trimmed, "\r\n")
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
	case "get":
		return []byte(store.Get(command[4])), nil
	case "set":
		if len(command) == 11 {
			ttl, err := strconv.ParseInt(command[10], 10, 64)
			if err != nil {
				return nil, err
			}
			return []byte(store.Set(command[4], command[6], ttl)), nil
		}
		return []byte(store.Set(command[4], command[6], 0)), nil
	}

	return nil, errors.New("command unsupported")
}

func handleConnection(conn net.Conn, storeService Store) {
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

		response, err := parser(data, storeService)
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

		storeService := NewStringStoreService()
		fmt.Println("Connection received, starting thread")
		go handleConnection(conn, storeService)
	}
}
