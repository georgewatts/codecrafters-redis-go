package main

import (
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	_ "embed"
)

//go:embed blank.rdb
var rdbHex string

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

	replicants []net.Conn
}

func (replication Replication) String() string {
	return fmt.Sprintf("role:%s\r\nmaster_replid:%s\r\nmaster_repl_offset:%d", replication.role, replication.id, replication.offset)
}

func (replication *Replication) RegisterReplicant(replicant net.Conn) {
	replication.replicants = append(replication.replicants, replicant)
}

func (replication *Replication) Propagate(data []string) {
	for i := 0; i < len(replication.replicants); i++ {
		replication.replicants[i].Write([]byte(NewRESPArray(data)))
	}
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

func Ping() []byte {
	return []byte(NewRESPArray([]string{"PING"}))
}

func ReplConfPort(port string) []byte {
	return []byte(NewRESPArray([]string{"REPLCONF", "listening-port", port}))
}

func ReplConfCapa() []byte {
	return []byte(NewRESPArray([]string{"REPLCONF", "capa", "psync2"}))
}

func ReplConfSync() []byte {
	return []byte(NewRESPArray([]string{"PSYNC", "?", "-1"}))
}

func ReplFullResync() []byte {
	decoded, err := hex.DecodeString(rdbHex)
	if err != nil {
		fmt.Println("Could not decode rdb", err)
	}

	return append([]byte(fmt.Sprintf("$%d\r\n", len(decoded))), decoded...)
}

func (replication Replication) Handshake(conn net.Conn, port string) {
	okResp := NewSimpleString("OK")
	pongResp := NewSimpleString("PONG")
	buf := make([]byte, len(pongResp))
	conn.Write(Ping())
	conn.Read(buf)
	conn.Write(ReplConfPort(port))
	buf = make([]byte, len(okResp))
	conn.Read(buf)
	conn.Write(ReplConfCapa())
	buf = make([]byte, len(okResp))
	conn.Read(buf)
	conn.Write(ReplConfSync())
}

func RespClient() {
}
