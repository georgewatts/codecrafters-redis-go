package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

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

		response := HandleData(data, storeService, replication)
		// if err != nil {
		// 	fmt.Println(err)
		// 	break
		// }
		conn.Write(response)
	}
}

func main() {
	var port string
	flag.StringVar(&port, "port", "6379", "Server Port")
	var replicaOf string
	flag.StringVar(&replicaOf, "replicaof", "master", "Replica of another instance")
	flag.Parse()

	replication := NewReplication("8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb", replicaOf, 0)
	if replication.role == SLAVE {
		masterConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", replication.host, replication.port))
		if err != nil {
			fmt.Println("Could not connect to master", err)
		}
		defer masterConn.Close()
		replication.Handshake(masterConn, port)
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
		go handleConnection(conn, storeService, replication)
	}
}
