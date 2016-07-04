package logde

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
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
}

func New(c Config) (*Server, error) {
	l, err := net.Listen("tcp", c.addr)
	if err != nil {
		return nil, err
	}

	server := &Server{
		l:       l,
		storage: c.storage,
	}

	return server, nil
}

func (s *Server) Run() error {
	fmt.Println("running")
	for {
		// Listen for an incoming connection.
		conn, err := s.l.Accept()
		if err != nil {
			return err
		}
		// Handle connections in a new goroutine.
		go s.handleRequest(&connection{conn: conn})
	}

	return nil
}

func (s *Server) Close() error {
	fmt.Println("closing")
	return s.l.Close()
}

func (s *Server) handleRequest(conn *connection) {
	for {
		request, err := Parse(conn)
		if err != nil {
			conn.Close()
			break
		}

		arguments := request.arguments

		switch request.command {
		case "QUIT":
			conn.Write(resultOK)
			conn.Close()
		case "SET":
			switch len(arguments) {
			case 3:

				ttl, err := strconv.Atoi(arguments[1])
				if err != nil || ttl < 0 {
					conn.Write(resultError)
					continue
				}

				dataLength, _ := strconv.Atoi(arguments[2])
				data, _ := request.data(dataLength)
				s.storage.Set(arguments[0], string(data), int64(ttl))
				conn.Write(resultOK)
			default:
				conn.Write(resultError)
				continue
			}
		case "GET":
			if len(arguments) != 1 {
				conn.Write(resultError)
				continue
			}
			value, err := s.storage.Get(string(arguments[0]))
			if err == nil {
				conn.WriteValues(value)
			} else {
				if err == ErrNotFound {
					conn.Write(resultNotFound)
				} else {
					conn.Write(resultError)
				}
			}
		case "HSET":
			switch len(arguments) {
			case 4:
				ttl, err := strconv.Atoi(arguments[2])
				if err != nil || ttl < 0 {
					conn.Write(resultError)
					continue
				}

				dataLength, _ := strconv.Atoi(request.arguments[3])
				data, _ := request.data(dataLength)

				s.storage.HSet(arguments[0], arguments[1], string(data), int64(ttl))
			default:
				conn.Write(resultError)
				continue
			}

			conn.Write(resultOK)

		case "HGET":
			if len(arguments) != 2 {
				conn.WriteError()
				continue
			}

			value, err := s.storage.HGet(arguments[0], arguments[1])
			if err == nil {
				conn.WriteValues(value)
			} else {
				if err == ErrNotFound {
					conn.Write(resultNotFound)
				} else {
					conn.Write(resultError)
				}
			}
		case "HGETALL":
			if len(arguments) != 1 {
				conn.WriteError()
			}

			if hash, err := s.storage.HGetAll(arguments[0]); err == nil {
				values := make([]string, len(hash)*2)
				i := 0
				for key, value := range hash {
					values[i] = key
					values[i+1] = value
					i += 2
				}

				conn.WriteValues(values...)
			}

			conn.WriteError()

		case "DELETE":
			if len(arguments) != 1 {
				conn.WriteError()
			}

			if err := s.storage.Delete(arguments[0]); err != nil {
				conn.WriteError()
			} else {
				conn.WriteOK()
			}

		case "KEYS":
			keys, err := s.storage.Keys()
			if err != nil {
				conn.Write(resultError)
			} else {
				conn.WriteValues(keys...)
			}
		default:
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
	buf := &bytes.Buffer{}

	fmt.Fprint(buf, string(resultValues))
	fmt.Fprintf(buf, "%d\r\n", len(values))
	for _, value := range values {
		fmt.Fprintf(buf, "%d\r\n", len(value))
		fmt.Fprintf(buf, "%s", value)
	}

	io.Copy(c.conn, buf)
}

func (c *connection) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

func (c *connection) Close() error {
	return c.conn.Close()
}
