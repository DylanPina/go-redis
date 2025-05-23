package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
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

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from connection: %s", err.Error())
			return
		}

		fmt.Printf("Received command: %s\n", line)
		conn.Write([]byte("+PONG\r\n"))
	}
}
