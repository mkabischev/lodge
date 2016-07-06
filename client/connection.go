package client

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/mkabischev/lodge/ioutil"
)

var (
	replyError        = "ERROR"
	replyOK           = "OK"
	replyValues       = "VALUES"
	replyNotFound     = "NOT_FOUND"
	replyAuthRequired = "AUTH_REQUIRED"
	replyBadFormat    = "BAD_FORMAT"

	ErrNotFound     = errors.New("Key not found")
	ErrSyntax       = errors.New("Syntax error")
	ErrServer       = errors.New("Server error")
	ErrAuthRequired = errors.New("Authentication required")
	ErrBadFormat    = errors.New("Bad format")
)

// connection is wrapper for net.Conn and contains logic about logde protocol.
// connection isn`t thread-safety. Each request must use separate connection (via pool).
type connection struct {
	c net.Conn
}

// newConnection returns new connection instance.
func newConnection(c net.Conn) *connection {
	return &connection{
		c: c,
	}
}

// send sends commands and arguments to server.
func (c *connection) send(operation string, args []interface{}, data interface{}) ([]string, error) {
	buf := bufio.NewWriter(c.c)

	buf.WriteString(operation)
	for _, arg := range args {
		buf.WriteString(fmt.Sprintf(" %v", arg))

	}
	buf.WriteString("\r\n")

	if data != nil {
		buf.WriteString(fmt.Sprintf("%v\r\n", data))
	}

	if err := buf.Flush(); err != nil {
		return nil, err
	}

	return c.parseResponse()
}

// parseResponse reads response from connection and then parses it.
func (c *connection) parseResponse() ([]string, error) {
	reader := bufio.NewReader(c.c)
	line, _, _ := reader.ReadLine()

	switch string(line) {
	case replyError:
		return nil, fmt.Errorf("some error")
	case replyOK:
		return nil, nil
	case replyValues:
		// reading next line containing number of values
		values, _, err := reader.ReadLine()
		if err != nil {
			return nil, fmt.Errorf("Error reading from response: %v", err)
		}
		valuesNumber, _ := strconv.Atoi(string(values))

		result := make([]string, valuesNumber)
		for i := 0; i < valuesNumber; i++ {
			value, err := c.readValue(reader)
			if err != nil {
				return nil, err
			}

			result[i] = string(value)
		}

		return result, nil
	case replyNotFound:
		return nil, ErrNotFound
	case replyAuthRequired:
		return nil, ErrAuthRequired
	case replyBadFormat:
		return nil, ErrBadFormat
	default:
		return nil, ErrServer
	}
}

// readValues reads next value from response
func (c *connection) readValue(r *bufio.Reader) ([]byte, error) {
	lengthLine, _, err := r.ReadLine()
	if err != nil {
		return nil, err
	}

	length, _ := strconv.Atoi(string(lengthLine))
	return ioutil.Read(r, length)
}
