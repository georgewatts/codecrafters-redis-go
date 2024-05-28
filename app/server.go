package main

import (
	"errors"
	"flag"
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

const (
	MASTER = "master"
	SLAVE  = "slave"
)

type Replication struct {
	role string
}

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

func parser(str string, store Store, replication Replication) ([]byte, error) {
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
	case "info":
		return []byte(NewBulkString(fmt.Sprintf("role:%s", replication.role))), nil
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

func handleConnection(conn net.Conn, storeService Store, replication Replication) {
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

		response, err := parser(data, storeService, replication)
		if err != nil {
			fmt.Println(err)
			break
		}
		conn.Write(response)
	}
}

func main() {
	var port string
	flag.StringVar(&port, "port", "6379", "Server Port")
	var replicaOf string
	flag.StringVar(&replicaOf, "replicaof", "master", "Replica of another instance")
	flag.Parse()

	replication := Replication{
		role: MASTER,
	}
	if strings.ToLower(replicaOf) != MASTER {
		replication.role = SLAVE
	}

	address := fmt.Sprintf("0.0.0.0:%s", port)
	l, err := net.Listen("tcp", address)
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
		go handleConnection(conn, storeService, replication)
	}
}
