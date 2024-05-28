package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

const (
	NULL_BULK_STRING = "$-1\r\n"
	OK               = "OK"
)

func newSimpleString(str string) string {
	return fmt.Sprintf("+%v\r\n", str)
}

func newBulkString(str string) string {
	if str == "" {
		return NULL_BULK_STRING
	}
	return fmt.Sprintf("$%d\r\n%v\r\n", len(str), str)
}

func pong() []byte {
	return []byte("+PONG\r\n")
}

func echo(str string) []byte {
	return []byte(newBulkString(str))
}

func parser(str string, store Store) ([]byte, error) {
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
	case "get":
		return []byte(store.Get(command[4])), nil
	case "set":
		return []byte(store.Set(command[4], command[6])), nil
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

type StringStoreService struct {
	store map[string]string
}

type Store interface {
	Set(string, string) string
	Get(string) string
}

func (storeService *StringStoreService) Get(key string) string {
	val := storeService.store[key]

	if val == "" {
		return newBulkString("")
	}

	return newBulkString(val)
}

func (storeService *StringStoreService) Set(key string, val string) string {
	storeService.store[key] = val

	return newBulkString(OK)
}

func newStringStoreService() *StringStoreService {
	return &StringStoreService{
		store: map[string]string{},
	}
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	storeService := newStringStoreService()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		fmt.Println("Connection received, starting thread")
		go handleConnection(conn, storeService)
	}
}
