// resp is a Redis Serialization Protocol (RESP) parser.
// See https://redis.io/docs/reference/protocol-spec/#resp-protocol-description
package redis

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// RESP protocol types
type (
	RESPType         any
	RESPSimpleString string
	RESPError        string
	RESPInteger      int64
	RESPBulkString   string
	RESPArray        []RESPType
)

// RESP protocol prefixes
const (
	RESPSimpleStringPrefix = '+'
	RESPErrorPrefix        = '-'
	RESPIntegerPrefix      = ':'
	RESPBulkStringPrefix   = '$'
	RESPArrayPrefix        = '*'
)

// RESP protocol commands
const (
	CommandPing   = "PING"
	CommandPong   = "PONG"
	CommandEcho   = "ECHO"
	CommandSet    = "SET"
	CommandGet    = "GET"
	CommandPx     = "PX"
	CommandConfig = "CONFIG"
)

// RESP protocol subcommands
const (
	// Config subcommands
	SubcommandConfigDir        = "dir"
	SubcommandConfigDBFileName = "dbfilename"
)

func Parse(reader *bufio.Reader) (RESPType, error) {
	prefix, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch prefix {
	case RESPSimpleStringPrefix:
		return parseSimpleString(reader)
	case RESPErrorPrefix:
		return parseError(reader)
	case RESPIntegerPrefix:
		return parseInteger(reader)
	case RESPBulkStringPrefix:
		return parseBulkString(reader)
	case RESPArrayPrefix:
		return parseArray(reader)
	default:
		return nil, fmt.Errorf("unknown RESP type: %c", prefix)
	}
}

func parseSimpleString(reader *bufio.Reader) (RESPSimpleString, error) {
	line, err := readRESPLine(reader)
	if err != nil {
		return "", err
	}
	return RESPSimpleString(line), nil
}

func parseError(reader *bufio.Reader) (RESPError, error) {
	line, err := readRESPLine(reader)
	if err != nil {
		return "", err
	}
	return RESPError(line), nil
}

func parseInteger(reader *bufio.Reader) (RESPInteger, error) {
	line, err := readRESPLine(reader)
	if err != nil {
		return 0, err
	}
	val, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return 0, err
	}
	return RESPInteger(val), nil
}

func parseBulkString(reader *bufio.Reader) (RESPBulkString, error) {
	line, err := readRESPLine(reader)
	if err != nil {
		return "", err
	}

	length, err := strconv.Atoi(line)
	if err != nil {
		return "", err
	}
	if length == -1 {
		return "", nil // null bulk string
	}

	buf := make([]byte, length+2) // include \r\n
	if _, err := io.ReadFull(reader, buf); err != nil {
		return "", err
	}

	return RESPBulkString(string(buf[:length])), nil
}

func parseArray(reader *bufio.Reader) (RESPArray, error) {
	line, err := readRESPLine(reader)
	if err != nil {
		return nil, err
	}

	count, err := strconv.Atoi(line)
	if err != nil {
		return nil, err
	}
	if count == -1 {
		return nil, nil
	}

	result := make(RESPArray, 0, count)
	for range count {
		elem, err := Parse(reader)
		if err != nil {
			return nil, err
		}
		result = append(result, elem)
	}
	return result, nil
}

func readRESPLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	if !strings.HasSuffix(line, "\r\n") {
		return "", errors.New("malformed line (missing \\r\\n)")
	}
	return strings.TrimSuffix(line, "\r\n"), nil
}
