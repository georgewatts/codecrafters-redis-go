package main

import (
	"flag"
	"fmt"
	"net"
	"sync"
)

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

	wg := sync.WaitGroup{}
	wg.Add(1)
	go RespServer(port, &replication)

	wg.Wait()
}
