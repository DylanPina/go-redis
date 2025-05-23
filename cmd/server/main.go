package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
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
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from connection: %s", err.Error())
		return
	}

	request := string(buf[:n])
	fmt.Printf("Received data: %s\n", request)

	if strings.TrimSpace(request) != "PING" {
		fmt.Fprintf(os.Stderr, "Expected PING, got: %s", request)
		conn.Write([]byte("-ERR unknown command\r\n"))
		return
	}

	conn.Write([]byte("+PONG\r\n"))
}
