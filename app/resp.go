package main

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// INCOMING COMMANDS
const (
	PING = "PING"
	ECHO = "ECHO"
	GET  = "GET"
	SET  = "SET"
	INFO = "INFO"
)

// OUTGOING COMMANDS
const (
	OK   = "OK"
	PONG = "PONG"
)

const NULL_BULK_STRING = "$-1\r\n"

type (
	ResponseSequence string
)

type ControlSequence struct {
	command string
	payload []string
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

func Parse(encodedSequence string) ControlSequence {
	trimmed := strings.TrimSpace(encodedSequence)
	tokens := strings.Split(trimmed, "\r\n")

	if tokens[0][0] != '*' {
		fmt.Println("Input was not an array sequence")
	}

	// Removes string length characters e.g. $9 (following word is 9 chars)
	// Also trims the first character which should be an asterisk denoting an array
	tokens = slices.DeleteFunc(tokens[1:], func(element string) bool {
		match, _ := regexp.MatchString(`\$\d+`, element)
		return match
	})
	tokens[0] = strings.ToUpper(tokens[0])

	controlSequence := ControlSequence{}

	switch tokens[0] {
	case PING:
		controlSequence.command = PING
	case INFO:
		controlSequence.command = INFO
	case ECHO:
		controlSequence.command = ECHO
		controlSequence.payload = tokens[1:]
	case GET:
		controlSequence.command = GET
		controlSequence.payload = tokens[1:]
	case SET:
		controlSequence.command = SET
		controlSequence.payload = tokens[1:]
	}

	return controlSequence
}

func (controlSequence ControlSequence) Execute(store Store, replication Replication) ResponseSequence {
	switch controlSequence.command {
	case PING:
		return NewSimpleString(PONG)
	case ECHO:
		return NewBulkString(controlSequence.payload[0])
	case GET:
		return NewBulkString(store.Get(controlSequence.payload[0]))
	case SET:
		fmt.Printf("controlSequence: %v\n", controlSequence)
		if len(controlSequence.payload) == 4 {
			ttl, err := strconv.ParseInt(controlSequence.payload[3], 10, 64)
			if err != nil {
				fmt.Println("Could not parse TTL int")
				return ""
			}
			store.Set(controlSequence.payload[0], controlSequence.payload[1], ttl)
		} else {
			store.Set(controlSequence.payload[0], controlSequence.payload[1], 0)
		}
		return NewBulkString(OK)
	case INFO:
		return NewBulkString(replication.String())
	}

	// TODO: not sure what to do here
	return ""
}

func (responseSequence ResponseSequence) Write() []byte {
	return []byte(responseSequence)
}

func HandleData(data string, store Store, replication Replication) []byte {
	controlSequence := Parse(data)
	response := controlSequence.Execute(store, replication)
	return response.Write()
}
