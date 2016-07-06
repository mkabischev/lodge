package server

import (
	"bufio"
	"fmt"
	"net"
)

var (
	crlf               = []byte("\r\n")
	resultOK           = []byte("OK\r\n")
	resultValues       = []byte("VALUES\r\n")
	resultError        = []byte("ERROR\r\n")
	resultWrongCommand = []byte("WRONG_COMMAND\r\n")
	resultAuthRequired = []byte("AUTH_REQUIRED\r\n")
	resultNotFound     = []byte("NOT_FOUND\r\n")
	resultBadFormat    = []byte("BAD_FORMAT\r\n")
)

type Server struct {
	storage Storage
	l       net.Listener
	users   *UserList

	commands map[string]command
}

func New(s Storage, users *UserList) *Server {
	server := &Server{
		storage: s,
		users:   users,
		commands: map[string]command{
			"GET":     getCommand{},
			"SET":     setCommand{},
			"HGET":    hGetCommand{},
			"HSET":    hSetCommand{},
			"HGETALL": hGetAllCommand{},
			"DELETE":  deleteCommand{},
			"KEYS":    keysCommand{},
			"EXPIRE":  expireCommand{},
		},
	}

	return server
}

func (s *Server) Serve(l net.Listener) error {
	s.l = l

	for {
		// Listen for an incoming connection.
		conn, err := s.l.Accept()
		if err != nil {
			return err
		}
		go s.handleRequest(&connection{conn: conn, authenticated: s.users == nil})
	}

	return nil
}

func (s *Server) ListenAndServe(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return s.Serve(l)
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

		// authentication checking
		if request.command == "AUTH" {
			if conn.authenticated {
				conn.WriteOK()
				continue
			}

			if len(request.arguments) != 2 {
				conn.WriteError()
			}
			if s.users.Validate(request.arguments[0], request.arguments[1]) {
				conn.WriteOK()
				conn.authenticated = true
			} else {
				conn.Write(resultAuthRequired)
			}
			continue
		}

		if cmd, ok := s.commands[request.command]; ok {
			if !conn.authenticated {
				conn.Write(resultAuthRequired)
				continue
			}
			if len(request.arguments) != cmd.arguments() {
				conn.WriteError()
				continue
			}

			values, err := cmd.process(request, s.storage)
			if err != nil {
				switch err {
				case errNotFound:
					conn.Write(resultNotFound)
				case errBadFormat:
					conn.Write(resultBadFormat)
				default:
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
			conn.Write(resultWrongCommand)
		}
	}
}

func (s *Server) authenticate(r *request, conn *connection) {
	if conn.authenticated {
		return
	}

	if s.users == nil {
		conn.authenticated = true
		return
	}

}

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
