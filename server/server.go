package server

import (
	"net"
	"time"
)

// Config is configuration for lodge server
type Config struct {
	// Is users specified then server will require authentication.
	Users *UserList
	// Timeout for queries processing.
	Timeout time.Duration
}

func DefaultConfig() *Config {
	return &Config{
		Timeout: 1 * time.Second,
	}
}

type Server struct {
	storage Storage
	l       net.Listener
	users   *UserList
	timeout time.Duration

	commands map[string]command
}

func New(s Storage, config *Config) *Server {
	server := &Server{
		storage: s,
		users:   config.Users,
		timeout: config.Timeout,
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

		go s.handleConnection(&connection{
			conn:          conn,
			authenticated: s.users == nil,
		})
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

func (s *Server) handleConnection(conn *connection) {
	for {
		request, err := Parse(conn)
		if err != nil {
			conn.Close()
			break
		}

		s.handleRequest(conn, request)
	}
}

func (s *Server) handleRequest(conn *connection, request *request) {
	// authentication checking
	if request.command == "AUTH" {
		if conn.authenticated {
			conn.WriteOK()
			return
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
		return
	}

	if cmd, ok := s.commands[request.command]; ok {
		if !conn.authenticated {
			conn.Write(resultAuthRequired)
			return
		}
		if len(request.arguments) != cmd.arguments() {
			conn.WriteError()
			return
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

			return
		}

		if len(values) == 0 {
			conn.WriteOK()
			return
		}

		conn.WriteValues(values...)
		return
	}

	conn.Write(resultWrongCommand)
}
