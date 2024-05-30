package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type ResponseData struct {
	response ResponseSequence
	upgraded bool
}

type Command interface {
	// Execute should validate and run the command
	Execute(store Store, replication *Replication) (ResponseData, error)
}

type PingCommand struct{}

func (pingCommand PingCommand) Execute(store Store, replication *Replication) (ResponseData, error) {
	return ResponseData{
		response: NewSimpleString(PONG),
	}, nil
}

type EchoCommand struct {
	response string
}

func (echoCommand EchoCommand) Execute(store Store, replication *Replication) (ResponseData, error) {
	return ResponseData{
		response: NewBulkString(echoCommand.response),
	}, nil
}

type GetCommand struct {
	key string
}

func (getCommand GetCommand) Execute(store Store, replication *Replication) (ResponseData, error) {
	return ResponseData{
		response: NewBulkString(store.Get(getCommand.key)),
	}, nil
}

type SetCommand struct {
	key string
	val string
	ttl int64
}

func (setCommand SetCommand) Execute(store Store, replication *Replication) (ResponseData, error) {
	store.Set(setCommand.key, setCommand.val, setCommand.ttl)
	replication.Propagate([]string{"SET", setCommand.key, setCommand.val})
	return ResponseData{
		response: NewBulkString(OK),
	}, nil
}

type InfoCommand struct{}

func (infoCommand InfoCommand) Execute(store Store, replication *Replication) (ResponseData, error) {
	return ResponseData{
		response: NewBulkString(replication.String()),
	}, nil
}

type ReplConfCommand struct {
	payload []string
}

func (r ReplConfCommand) Execute(store Store, replication *Replication) (ResponseData, error) {
	if r.payload[0] == "listening-port" {
		// replication.RegisterReplicant(r.payload[1])
		return ResponseData{response: NewSimpleString(OK), upgraded: true}, nil
	}
	return ResponseData{response: NewSimpleString(OK)}, nil
}

type PsyncCommand struct{}

func (p PsyncCommand) Execute(store Store, replication *Replication) (ResponseData, error) {
	// TODO: This could probs be better
	return ResponseData{response: NewSimpleString(fmt.Sprintf("FULLRESYNC %s %d", replication.id, replication.offset)) + ResponseSequence(ReplFullResync())}, nil
}

func parse(encodedSequence string) (Command, error) {
	trimmed := strings.TrimSpace(encodedSequence)
	tokens := strings.Split(trimmed, "\r\n")

	if tokens[0][0] != '*' {
		return nil, errors.New("input was not an array sequence")
	}

	// Removes string length characters e.g. $9 (following word is 9 chars)
	// Also trims the first character which should be an asterisk denoting an array
	tokens = slices.DeleteFunc(tokens[1:], func(element string) bool {
		match, _ := regexp.MatchString(`\$\d+`, element)
		return match
	})
	tokens[0] = strings.ToUpper(tokens[0])

	switch tokens[0] {
	case PING:
		return PingCommand{}, nil
	case ECHO:
		return EchoCommand{response: tokens[1]}, nil
	case GET:
		return GetCommand{key: tokens[1]}, nil
	case SET:
		if len(tokens) == 5 {
			ttl, err := strconv.ParseInt(tokens[4], 10, 64)
			if err != nil {
				return nil, err
			}
			return SetCommand{key: tokens[1], val: tokens[2], ttl: ttl}, nil
		}
		return SetCommand{key: tokens[1], val: tokens[2]}, nil
	case INFO:
		return InfoCommand{}, nil
	case REPLCONF:
		return ReplConfCommand{payload: tokens[1:]}, nil
	case PSYNC:
		return PsyncCommand{}, nil
	}
	return nil, errors.New("unrecognised command " + tokens[0])
}

func handleData(data string, store Store, replication *Replication) (ResponseData, error) {
	command, err := parse(data)
	if err != nil {
		return ResponseData{}, err
	}
	response, err := command.Execute(store, replication)
	if err != nil {
		return ResponseData{}, err
	}
	return response, nil
}

func handleConnection(conn net.Conn, storeService Store, replication *Replication) {
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

		response, err := handleData(data, storeService, replication)
		if err != nil {
			fmt.Println(err)
			break
		}

		if response.upgraded {
			replication.RegisterReplicant(conn)
		}

		conn.Write([]byte(response.response))
	}
}

func RespServer(port string, replication *Replication) {
	address := fmt.Sprintf("0.0.0.0:%s", port)
	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Failed to bind to port ", port)
		os.Exit(1)
	}
	defer l.Close()

	storeService := NewStringStoreService()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		// fmt.Printf("conn.RemoteAddr().String(): %v\n", conn.RemoteAddr().String())
		// replication.RegisterReplicant(conn.RemoteAddr().String())
		go handleConnection(conn, storeService, replication)
	}
}
