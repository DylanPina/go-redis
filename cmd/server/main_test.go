package main

import (
	"context"
	"flag"
	"os"
	"strconv"
	"testing"
	"time"

	redisServer "github.com/DylanPina/go-redis/internal/redis"
	redisClient "github.com/redis/go-redis/v9"
)

var (
	testPort       = flag.Int("port", 6379, "Port that the DNS server to listen on (default: 6379)")
	testDir        = flag.String("dir", "", "Directory where the Redis server is running (default: current directory)")
	testDBFileName = flag.String("dbfilename", "dump.rdb", "Name of the Redis database file (default: dump.rdb)")
)

// TestMain is the entry point for the test suite
func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

// TestPing tests the Redis server's ability to respond to a PING command
func TestPong(t *testing.T) {
	rdb := createClient()

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
	rdb := createClient()

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
	rdb := createClient()

	// Set a key-value pair in the server
	ctx := context.Background()
	key := "testKey"
	value := "testValue"
	err := rdb.Do(ctx, redisServer.CommandSet, key, value).Err()
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

// TestSetGetNonExistentKey tests the Redis server's ability to handle non-existent keys
func TestSetGetNonExistentKey(t *testing.T) {
	rdb := createClient()

	// Try to get a non-existent key
	ctx := context.Background()
	key := "nonExistentKey"
	gotValue, err := rdb.Do(ctx, redisServer.CommandGet, key).Result()
	if err == nil {
		t.Fatalf("Expected error for non-existent key, got value: %s", gotValue)
	}

	if gotValue != nil {
		t.Fatalf("Expected nil for non-existent key, got: %s", gotValue)
	}

	t.Logf("Get non-existent key successful, no value found for '%s'", key)
}

// TestSetGetWithExpiredKey tests the Redis server's ability to handle key expiration
func TestSetGetWithExpiredKey(t *testing.T) {
	rdb := createClient()

	// Set a key-value pair with expiration
	ctx := context.Background()
	key := "tempKey"
	value := "tempValue"
	expiration := 100 // 0.1 seconds in milliseconds
	err := rdb.Do(ctx, redisServer.CommandSet, key, value, redisServer.CommandPx, expiration).Err()
	if err != nil {
		t.Fatalf("Failed to set key with expiration: %v", err)
	}

	// Wait for the key to expire
	time.Sleep(1000 * time.Millisecond)

	// Try to get the value back from the server
	gotValue, err := rdb.Do(ctx, redisServer.CommandGet, key).Result()
	if err == nil {
		t.Fatalf("Expected key to be expired, but got value: %s", gotValue)
	}

	t.Logf("Set with expiration successful, key '%s' is expired as expected", key)
}

// TestSetGetWithUnexpiredKey tests the Redis server's ability to handle keys that should not expire
func TestSetGetWithUnexpiredKey(t *testing.T) {
	rdb := createClient()

	// Set a key-value pair without expiration
	ctx := context.Background()
	key := "permanentKey"
	value := "permanentValue"
	expiration := 100000 // 100 seconds in milliseconds
	err := rdb.Do(ctx, redisServer.CommandSet, key, value, redisServer.CommandPx, expiration).Err()
	if err != nil {
		t.Fatalf("Failed to set key without expiration: %v", err)
	}

	// Get the value back from the server
	gotValue, err := rdb.Do(ctx, redisServer.CommandGet, key).Result()
	if err != nil {
		t.Fatalf("Failed to get key: %v", err)
	}

	if gotValue != value {
		t.Fatalf("Expected value '%s', got: %s", value, gotValue)
	}

	t.Logf("Set and Get without expiration successful: %s = %s", key, gotValue)
}

func TestConfigSetGet(t *testing.T) {
	// Create a new Redis rdb
	rdb := createClient()

	// Set configuration directory
	ctx := context.Background()
	targetDir := *testDir
	err := rdb.Do(ctx, redisServer.CommandConfig, redisServer.CommandSet, redisServer.SubcommandConfigDir, targetDir).Err()
	if err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	// Get the directory value back
	gotDir, err := rdb.Do(ctx, redisServer.CommandConfig, redisServer.CommandGet, redisServer.SubcommandConfigDir).Result()
	if err != nil {
		t.Fatalf("Failed to get config directory: %v", err)
	}

	if gotDir != targetDir {
		t.Fatalf("Expected config directory value '%s', got: %s", targetDir, gotDir)
	}

	// Set db file name
	targetDBFileName := *testDBFileName
	err = rdb.Do(ctx, redisServer.CommandConfig, redisServer.CommandSet, redisServer.SubcommandConfigDBFileName, targetDBFileName).Err()
	if err != nil {
		t.Fatalf("Failed to set config db filename: %v", err)
	}

	// Get the db file name value back
	gotDbFileName, err := rdb.Do(ctx, redisServer.CommandConfig, redisServer.CommandGet, redisServer.SubcommandConfigDBFileName).Result()
	if err != nil {
		t.Fatalf("Failed to get config db filename: %v", err)
	}

	t.Logf("Config Set and Get successful: %s = %s", targetDBFileName, gotDbFileName)
}

func createClient() *redisClient.Client {
	return redisClient.NewClient(&redisClient.Options{
		Addr: "localhost:" + strconv.Itoa(*testPort),
	})
}
