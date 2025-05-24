// resp is a Redis Serialization Protocol (RESP) parser.
// See https://redis.io/docs/reference/protocol-spec/#resp-protocol-description
package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type (
	// RESP protocol types
	RespType         any
	RespSimpleString string
	RespError        string
	RespInteger      int64
	RespBulkString   string
	RespArray        []RespType
)

const (
	// RESP protocol prefixes
	RespSimpleStringPrefix = '+'
	RespErrorPrefix        = '-'
	RespIntegerPrefix      = ':'
	RespBulkStringPrefix   = '$'
	RespArrayPrefix        = '*'
)

const (
	// RESP protocol commands
	RespCommandPing = "PING"
	RespCommandEcho = "ECHO"
	RespCommandSet  = "SET"
	RespCommandGet  = "GET"
)

func Parse(reader *bufio.Reader) (RespType, error) {
	prefix, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch prefix {
	case RespSimpleStringPrefix:
		return parseSimpleString(reader)
	case RespErrorPrefix:
		return parseError(reader)
	case RespIntegerPrefix:
		return parseInteger(reader)
	case RespBulkStringPrefix:
		return parseBulkString(reader)
	case RespArrayPrefix:
		return parseArray(reader)
	default:
		return nil, fmt.Errorf("unknown RESP type: %c", prefix)
	}
}

func parseSimpleString(reader *bufio.Reader) (RespSimpleString, error) {
	line, err := readRespLine(reader)
	if err != nil {
		return "", err
	}
	return RespSimpleString(line), nil
}

func parseError(reader *bufio.Reader) (RespError, error) {
	line, err := readRespLine(reader)
	if err != nil {
		return "", err
	}
	return RespError(line), nil
}

func parseInteger(reader *bufio.Reader) (RespInteger, error) {
	line, err := readRespLine(reader)
	if err != nil {
		return 0, err
	}
	val, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return 0, err
	}
	return RespInteger(val), nil
}

func parseBulkString(reader *bufio.Reader) (RespBulkString, error) {
	line, err := readRespLine(reader)
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

	return RespBulkString(string(buf[:length])), nil
}

func parseArray(reader *bufio.Reader) (RespArray, error) {
	line, err := readRespLine(reader)
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

	result := make(RespArray, 0, count)
	for range count {
		elem, err := Parse(reader)
		if err != nil {
			return nil, err
		}
		result = append(result, elem)
	}
	return result, nil
}

func readRespLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	if !strings.HasSuffix(line, "\r\n") {
		return "", errors.New("malformed line (missing \\r\\n)")
	}
	return strings.TrimSuffix(line, "\r\n"), nil
}
