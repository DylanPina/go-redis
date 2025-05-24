package main

import (
	"flag"
	"io"
	"net"
	"os"
	"strconv"
	"testing"
)

var TEST_PORT = flag.Int("port", 6379, "Port that the DNS server to listen on (default: 6379)")

// TestMain is the entry point for the test suite
func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

// TestPing tests the ping command
func TestPong(t *testing.T) {
	// Simulate a connection to the server
	conn, err := net.Dial("tcp", "localhost:"+strconv.Itoa(*TEST_PORT))
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Send a PING command
	_, err = conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	if err != nil {
		t.Fatalf("Failed to send PING command: %v", err)
	}

	// Listen for PONG response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("Error reading from connection: %v", err)
	}

	t.Logf("Received response: %s", string(buf[:n]))

	if string(buf[:n]) != "+PONG\r\n" {
		t.Fatalf("Expected PONG, got: %s", string(buf[:n]))
	}
}
