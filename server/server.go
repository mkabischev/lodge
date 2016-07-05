package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

var (
	crlf           = []byte("\r\n")
	resultOK       = []byte("OK\r\n")
	resultValues   = []byte("VALUES\r\n")
	resultError    = []byte("ERROR\r\n")
	resultNotFound = []byte("NOT_FOUND\r\n")
)

type Config struct {
	addr    string
	storage Storage
}

func DefaultConfig() Config {
	return Config{
		addr:    ":20000",
		storage: NewMemory(),
	}
}

func (c Config) WithAddr(addr string) Config {
	res := c
	res.addr = addr

	return res
}

func (c Config) WithStorage(storage Storage) Config {
	res := c
	res.storage = storage

	return res
}

type Server struct {
	addr    string
	storage Storage
	l       net.Listener

	commands map[string]command
}

func New(c Config) (*Server, error) {
	server := &Server{
		addr:    c.addr,
		storage: c.storage,
		commands: map[string]command{
			"GET":     getCommand{},
			"SET":     setCommand{},
			"HGET":    hGetCommand{},
			"HSET":    hSetCommand{},
			"HGETALL": hGetAllCommand{},
			"DELETE":  deleteCommand{},
			"KEYS":    keysCommand{},
		},
	}

	return server, nil
}

func (s *Server) Run() error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	s.l = l

	log.Printf("listening on: %v", s.addr)
	for {
		// Listen for an incoming connection.
		conn, err := s.l.Accept()
		if err != nil {
			return err
		}
		go s.handleRequest(&connection{conn: conn})
	}

	return nil
}

func (s *Server) Close() error {
	return s.l.Close()
}

func (s *Server) handleRequest(conn *connection) {
	for {
		request, err := Parse(conn)
		if err != nil {
			conn.Close()
			break
		}

		if cmd, ok := s.commands[request.command]; ok {
			if len(request.arguments) != cmd.arguments() {
				conn.WriteError()
				continue
			}

			values, err := cmd.process(request, s.storage)
			if err != nil {
				if err == errNotFound {
					conn.Write(resultNotFound)
				} else {
					conn.WriteError()
				}

				continue
			}

			if len(values) == 0 {
				conn.WriteOK()
				continue
			}

			conn.WriteValues(values...)
		} else {
			fmt.Fprintf(conn, "unknown command: %v\n", request.command)
		}
	}
}

type connection struct {
	conn net.Conn
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
