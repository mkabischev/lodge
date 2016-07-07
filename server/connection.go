package server

import (
	"bufio"
	"fmt"
	"net"
)

var (
	resultOK           = []byte("OK\r\n")
	resultValues       = []byte("VALUES\r\n")
	resultError        = []byte("ERROR\r\n")
	resultWrongCommand = []byte("WRONG_COMMAND\r\n")
	resultAuthRequired = []byte("AUTH_REQUIRED\r\n")
	resultNotFound     = []byte("NOT_FOUND\r\n")
	resultBadFormat    = []byte("BAD_FORMAT\r\n")
)

type connection struct {
	conn          net.Conn
	authenticated bool
}

func (c *connection) WriteOK() {
	c.conn.Write(resultOK)
}

func (c *connection) WriteError() {
	c.conn.Write(resultError)
}

func (c *connection) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

func (c *connection) WriteValues(values ...string) {
	w := bufio.NewWriter(c)

	fmt.Fprintf(w, "%s", resultValues)
	fmt.Fprintf(w, "%d\r\n", len(values))
	for _, value := range values {
		fmt.Fprintf(w, "%d\r\n", len(value))
		fmt.Fprintf(w, "%s", value)
	}

	w.Flush()
}

func (c *connection) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

func (c *connection) Close() error {
	return c.conn.Close()
}
