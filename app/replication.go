package main

import (
	"fmt"
	"net"
	"strings"
)

const (
	MASTER = "master"
	SLAVE  = "slave"
)

type Replication struct {
	id     string
	role   string
	host   string
	port   string
	offset int
}

func (replication Replication) String() string {
	return fmt.Sprintf("role:%s\r\nmaster_replid:%s\r\nmaster_repl_offset:%d", replication.role, replication.id, replication.offset)
}

func NewReplication(id string, address string, offset int) Replication {
	if address != MASTER {
		addressTokens := strings.Split(address, " ")
		return Replication{
			id:     id,
			role:   SLAVE,
			host:   addressTokens[0],
			port:   addressTokens[1],
			offset: offset,
		}
	}

	return Replication{
		id:     id,
		role:   MASTER,
		offset: offset,
	}
}

func ArrayString(elements []string) string {
	builder := strings.Builder{}
	builder.WriteByte('*')
	builder.WriteString(fmt.Sprint(len(elements)))
	builder.WriteString("\r\n")

	for _, v := range elements {
		builder.WriteByte('$')
		builder.WriteString(fmt.Sprint(len(v)))
		builder.WriteString("\r\n")
		builder.WriteString(v)
		builder.WriteString("\r\n")
	}

	return builder.String()
}

func (replication Replication) Ping() {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", replication.host, replication.port))
	if err != nil {
		fmt.Println("Could not connect to master", err)
	}
	defer conn.Close()

	conn.Write([]byte(ArrayString([]string{"PING"})))
}
