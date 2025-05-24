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
		val, err := resp.Parse(reader)
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "Error parsing RESP: %v\n", err)
			}
			return
		}

		req, ok := val.(resp.RESPArray)
		if !ok {
			fmt.Fprintf(os.Stderr, "Expected RESP array, got %T\n", val)
			writeRESPError(conn, fmt.Errorf("expected RESP array, got %T", val))
			return
		}

		handleRESP(conn, req)
	}
}

func handleRESP(conn net.Conn, req resp.RESPArray) {
	if len(req) == 0 {
		writeRESPError(conn, fmt.Errorf("empty request"))
		return
	}

	cmd, ok := req[0].(resp.RESPBulkString)
	if !ok {
		writeRESPError(conn, fmt.Errorf("first element is not a bulk string"))
		return
	}

	switch cmd {
	case resp.CommandPing:
		handlePingCommand(conn)

	case resp.CommandEcho:
		handleEchoCommand(conn, req)

	default:
		handleUnknownCommand(conn, cmd)
	}
}

func handlePingCommand(conn net.Conn) {
	writeRESPSimpleString(conn, resp.CommandPong)
}

func handleEchoCommand(conn net.Conn, req resp.RESPArray) {
	if len(req) < 2 {
		writeRESPError(conn, fmt.Errorf("ECHO command requires a message"))
		return
	}

	msg, ok := req[1].(resp.RESPBulkString)

	if !ok {
		writeRESPError(conn, fmt.Errorf("second element is not a bulk string"))
		return
	}

	writeRESPBulkString(conn, msg)
}

func handleUnknownCommand(conn net.Conn, cmd resp.RESPBulkString) {
	writeRESPError(conn, fmt.Errorf("unknown command: %s", cmd))
}

func writeRESPError(conn net.Conn, err error) {
	_, err = fmt.Fprintf(conn, "-ERR %s\r\n", err.Error())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to connection: %s\n", err.Error())
	}
}

func writeRESPSimpleString(conn net.Conn, msg resp.RESPSimpleString) {
	_, err := fmt.Fprintf(conn, "+%s\r\n", msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to connection: %s\n", err.Error())
	}
}

func writeRESPBulkString(conn net.Conn, msg resp.RESPBulkString) {
	_, err := fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(msg), msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to connection: %s\n", err.Error())
	}
}
