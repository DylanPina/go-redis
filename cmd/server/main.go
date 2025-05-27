package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/DylanPina/go-redis/internal/redis"
)

var (
	port       = flag.Int("port", 6379, "Port that the DNS server to listen on (default: 6379)")
	dir        = flag.String("dir", "", "Directory where the Redis server is running (default: current directory)")
	dbFileName = flag.String("dbfilename", "dump.rdb", "Name of the Redis database file (default: dump.rdb)")
)

func main() {
	flag.Parse()

	redis.SetDirectory(*dir)
	redis.SetDBFileName(*dbFileName)

	fmt.Printf("Starting Redis server on port %d\n", *port)
	fmt.Printf("Using directory: %s\n", redis.GetDirectory())
	fmt.Printf("Using database file: %s\n", redis.GetDBFileName())

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

		fmt.Printf("Received request: %v\n", req)
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

	case redis.CommandConfig:
		handleConfigCommand(conn, req)

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

	expiration := int64(-1) // Default to no expiration
	if len(req) == 5 {
		pxStr, ok := req[3].(redis.RESPBulkString)
		if !ok {
			writeRESPError(conn, fmt.Errorf("fourth element is not a bulk string"))
			return
		}
		if strings.ToUpper(string(pxStr)) != redis.CommandPx {
			writeRESPError(conn, fmt.Errorf("fourth element must be 'PX' for expiration"))
			return
		}

		expirationStr, ok := req[4].(redis.RESPBulkString)
		if !ok {
			writeRESPError(conn, fmt.Errorf("fifth element is not a bulk string (expiration)"))
			return
		}

		exp, err := strconv.ParseInt(string(expirationStr), 10, 64)
		if err != nil {
			writeRESPError(conn, fmt.Errorf("invalid expiration value: %s", err.Error()))
			return
		}

		if exp < 0 {
			writeRESPError(conn, fmt.Errorf("expiration must be a non-negative integer"))
			return
		}

		expiration = exp
	}

	redis.Set(string(key), string(val), expiration)
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

	val, ok := redis.Get(string(key))
	if !ok {
		writeRESPNullBulkString(conn)
		return
	}

	writeRESPBulkString(conn, redis.RESPBulkString(val))
}

func handleConfigCommand(conn net.Conn, req redis.RESPArray) {
	if len(req) < 2 {
		writeRESPError(conn, fmt.Errorf("CONFIG command requires a subcommand"))
		return
	}

	subCmd, ok := req[1].(redis.RESPBulkString)
	if !ok {
		writeRESPError(conn, fmt.Errorf("second element is not a bulk string (subcommand)"))
		return
	}

	switch subCmd {
	case redis.CommandGet:
		handleConfigGetCommand(conn, req)

	case redis.CommandSet:
		handleConfigSetCommand(conn, req)

	default:
		writeRESPError(conn, fmt.Errorf("unknown CONFIG subcommand: %s", subCmd))
	}
}

func handleConfigGetCommand(conn net.Conn, req redis.RESPArray) {
	if len(req) != 3 {
		writeRESPError(conn, fmt.Errorf("CONFIG GET command requires a parameter"))
		return
	}

	param, ok := req[2].(redis.RESPBulkString)
	if !ok {
		writeRESPError(conn, fmt.Errorf("third element is not a bulk string (parameter)"))
		return
	}

	switch string(param) {
	case redis.SubcommandConfigDir:
		writeRESPBulkString(conn, redis.RESPBulkString(redis.GetDirectory()))
	case redis.SubcommandConfigDBFileName:
		writeRESPBulkString(conn, redis.RESPBulkString(redis.GetDBFileName()))
	default:
		writeRESPNullBulkString(conn)
	}
}

func handleConfigSetCommand(conn net.Conn, req redis.RESPArray) {
	if len(req) != 4 {
		writeRESPError(conn, fmt.Errorf("CONFIG SET command requires a parameter and value"))
		return
	}

	param, ok := req[2].(redis.RESPBulkString)
	if !ok {
		writeRESPError(conn, fmt.Errorf("third element is not a bulk string (parameter)"))
		return
	}

	value, ok := req[3].(redis.RESPBulkString)
	if !ok {
		writeRESPError(conn, fmt.Errorf("fourth element is not a bulk string (value)"))
		return
	}

	switch string(param) {
	case redis.SubcommandConfigDir:
		redis.SetDirectory(string(value))
	case redis.SubcommandConfigDBFileName:
		redis.SetDBFileName(string(value))
	default:
		writeRESPError(conn, fmt.Errorf("unknown CONFIG parameter: %s", param))
		return
	}

	writeRESPSimpleString(conn, "OK")
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
