package main

import (
	"context"
	"flag"
	"os"
	"strconv"
	"testing"

	redisServer "github.com/DylanPina/go-redis/internal/redis"
	redisClient "github.com/redis/go-redis/v9"
)

var TEST_PORT = flag.Int("port", 6379, "Port that the DNS server to listen on (default: 6379)")

// TestMain is the entry point for the test suite
func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

// TestPing tests the Redis server's ability to respond to a PING command
func TestPong(t *testing.T) {
	// Create a new Redis rdb
	rdb := redisClient.NewClient(&redisClient.Options{
		Addr: "localhost:" + strconv.Itoa(*TEST_PORT),
	})

	// Ping the server to check if it's running
	ctx := context.Background()
	pong, err := rdb.Do(ctx, redisServer.CommandPing).Result()
	if err != nil {
		t.Fatalf("Failed to ping Redis server: %v", err)
	}

	if pong != redisServer.CommandPong {
		t.Fatalf("Expected PONG, got: %s", pong)
	}

	t.Logf("Ping successful: %s", pong)
}

// TestEcho tests the Redis server's ability to respond to an ECHO command
func TestEcho(t *testing.T) {
	// Create a new Redis rdb
	rdb := redisClient.NewClient(&redisClient.Options{
		Addr: "localhost:" + strconv.Itoa(*TEST_PORT),
	})

	// Echo a message to the server
	ctx := context.Background()
	message := "Hello, Redis!"
	echo, err := rdb.Do(ctx, redisServer.CommandEcho, message).Result()
	if err != nil {
		t.Fatalf("Failed to echo message: %v", err)
	}

	if echo != message {
		t.Fatalf("Expected echo '%s', got: %s", message, echo)
	}

	t.Logf("Echo successful: %s", echo)
}

// TestSetGet tests the Redis server's ability to set and get a key-value pair
func TestSetGet(t *testing.T) {
	// Create a new Redis rdb
	rdb := redisClient.NewClient(&redisClient.Options{
		Addr: "localhost:" + strconv.Itoa(*TEST_PORT),
	})

	// Set a key-value pair in the server
	ctx := context.Background()
	key := "testKey"
	value := "testValue"
	err := rdb.Do(ctx, redisServer.CommandSet, key, value, 0).Err()
	if err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	// Get the value back from the server
	gotValue, err := rdb.Do(ctx, redisServer.CommandGet, key).Result()
	if err != nil {
		t.Fatalf("Failed to get key: %v", err)
	}

	if gotValue != value {
		t.Fatalf("Expected value '%s', got: %s", value, gotValue)
	}

	t.Logf("Set and Get successful: %s = %s", key, gotValue)
}
