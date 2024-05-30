package main

import (
	"fmt"
	"strings"
)

// INCOMING COMMANDS
const (
	PING     = "PING"
	ECHO     = "ECHO"
	GET      = "GET"
	SET      = "SET"
	INFO     = "INFO"
	REPLCONF = "REPLCONF"
	PSYNC    = "PSYNC"
)

// OUTGOING COMMANDS
const (
	OK   = "OK"
	PONG = "PONG"
)

const NULL_BULK_STRING = "$-1\r\n"

type ResponseSequence string

func (responseSequence ResponseSequence) Write() []byte {
	return []byte(responseSequence)
}

func NewSimpleString(str string) ResponseSequence {
	return ResponseSequence(fmt.Sprintf("+%v\r\n", str))
}

func NewBulkString(str string) ResponseSequence {
	if str == "" {
		return NULL_BULK_STRING
	}
	return ResponseSequence(fmt.Sprintf("$%d\r\n%v\r\n", len(str), str))
}

func NewRESPArray(elements []string) string {
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
