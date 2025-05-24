package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/DylanPina/go-redis/internal/redis"
)

func main() {
	port := flag.Int("port", 6379, "Port that the DNS server to listen on (default: 6379)")
	flag.Parse()

	l, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(*port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bind to port %d: %s\n", *port, err.Error())
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accepting connection: %s", err.Error())
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		val, err := redis.Parse(reader)
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "Error parsing RESP: %v\n", err)
			}
			return
		}

		req, ok := val.(redis.RESPArray)
		if !ok {
			fmt.Fprintf(os.Stderr, "Expected RESP array, got %T\n", val)
			writeRESPError(conn, fmt.Errorf("expected RESP array, got %T", val))
			return
		}

		handleRESP(conn, req)
	}
}

func handleRESP(conn net.Conn, req redis.RESPArray) {
	if len(req) == 0 {
		writeRESPError(conn, fmt.Errorf("empty request"))
		return
	}

	cmd, ok := req[0].(redis.RESPBulkString)
	if !ok {
		writeRESPError(conn, fmt.Errorf("first element is not a bulk string"))
		return
	}

	switch cmd {
	case redis.CommandPing:
		handlePingCommand(conn)

	case redis.CommandEcho:
		handleEchoCommand(conn, req)

	case redis.CommandSet:
		handleSetCommand(conn, req)

	case redis.CommandGet:
		handleGetCommand(conn, req)

	default:
		handleUnknownCommand(conn, cmd)
	}
}

func handlePingCommand(conn net.Conn) {
	writeRESPSimpleString(conn, redis.CommandPong)
}

func handleEchoCommand(conn net.Conn, req redis.RESPArray) {
	if len(req) < 2 {
		writeRESPError(conn, fmt.Errorf("ECHO command requires a message"))
		return
	}

	msg, ok := req[1].(redis.RESPBulkString)

	if !ok {
		writeRESPError(conn, fmt.Errorf("second element is not a bulk string"))
		return
	}

	writeRESPBulkString(conn, msg)
}

func handleSetCommand(conn net.Conn, req redis.RESPArray) {
	if len(req) < 3 {
		writeRESPError(conn, fmt.Errorf("SET command requires a key and value"))
		return
	}

	key, ok := req[1].(redis.RESPBulkString)
	if !ok {
		writeRESPError(conn, fmt.Errorf("second element is not a bulk string (key)"))
		return
	}

	val, ok := req[2].(redis.RESPBulkString)
	if !ok {
		writeRESPNullBulkString(conn)
		return
	}

	redis.Set(string(key), string(val))
	writeRESPSimpleString(conn, "OK")
}

func handleGetCommand(conn net.Conn, req redis.RESPArray) {
	if len(req) < 2 {
		writeRESPError(conn, fmt.Errorf("GET command requires a key"))
		return
	}
	if len(req) > 2 {
		writeRESPError(conn, fmt.Errorf("GET command takes only one argument (key)"))
		return
	}

	key, ok := req[1].(redis.RESPBulkString)
	if !ok {
		writeRESPError(conn, fmt.Errorf("second element is not a bulk string (key)"))
		return
	}

	val, exists := redis.Get(string(key))
	if !exists {
		writeRESPNullBulkString(conn)
		return
	}

	writeRESPBulkString(conn, redis.RESPBulkString(val))
}

func handleUnknownCommand(conn net.Conn, cmd redis.RESPBulkString) {
	writeRESPError(conn, fmt.Errorf("unknown command: %s", cmd))
}

func writeRESPError(conn net.Conn, err error) {
	_, err = fmt.Fprintf(conn, "-ERR %s\r\n", err.Error())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to connection: %s\n", err.Error())
	}
}

func writeRESPSimpleString(conn net.Conn, msg redis.RESPSimpleString) {
	_, err := fmt.Fprintf(conn, "+%s\r\n", msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to connection: %s\n", err.Error())
	}
}

func writeRESPBulkString(conn net.Conn, msg redis.RESPBulkString) {
	_, err := fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(msg), msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to connection: %s\n", err.Error())
	}
}

func writeRESPNullBulkString(conn net.Conn) {
	_, err := fmt.Fprint(conn, "$-1\r\n")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to connection: %s\n", err.Error())
	}
}
