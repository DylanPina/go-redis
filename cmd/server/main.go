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

		req, ok := val.(resp.RespArray)
		if !ok {
			fmt.Fprintf(os.Stderr, "Expected RESP array, got %T\n", val)
			fmt.Fprint(conn, "-ERR Expected RESP array\r\n")
			continue
		}

		handleResp(conn, req)
	}
}

func handleResp(conn net.Conn, req resp.RespArray) {
	if len(req) == 0 {
		writeRespError(conn, fmt.Errorf("empty request"))
		return
	}

	cmd, ok := req[0].(resp.RespBulkString)
	if !ok {
		writeRespError(conn, fmt.Errorf("first element is not a bulk string"))
		return
	}

	switch cmdStr := string(cmd); cmdStr {
	case resp.RespCommandPing:
		handlePingCommand(conn)

	case resp.RespCommandEcho:
		handleEchoCommand(conn, req)

	default:
		handleUnknownCommand(conn, cmd)
	}
}

func handlePingCommand(conn net.Conn) {
	writeRespSimpleString(conn, "PONG")
}

func handleEchoCommand(conn net.Conn, req resp.RespArray) {
	if len(req) < 2 {
		writeRespError(conn, fmt.Errorf("ECHO command requires a message"))
		return
	}

	msg, ok := req[1].(resp.RespBulkString)
	if !ok {
		writeRespError(conn, fmt.Errorf("second element is not a bulk string"))
		return
	}

	fmt.Fprintf(conn, "+%s\r\n", msg)
}

func handleUnknownCommand(conn net.Conn, cmd resp.RespBulkString) {
	writeRespError(conn, fmt.Errorf("unknown command: %s", cmd))
}

func writeRespError(conn net.Conn, err error) {
	_, err = fmt.Fprintf(conn, "-ERR %s\r\n", err.Error())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to connection: %s\n", err.Error())
	}
}

func writeRespSimpleString(conn net.Conn, msg string) {
	_, err := fmt.Fprintf(conn, "+%s\r\n", msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to connection: %s\n", err.Error())
	}
}
