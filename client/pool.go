package client

import (
	"net"
	"time"
)

type pool struct {
	addr string
	free chan net.Conn
}

// newPool
func newPool(addr string, size int) *pool {
	return &pool{
		addr: addr,
		free: make(chan net.Conn, size),
	}
}

// get returns free connection from the pool. If pool is empty then new connection will be created
func (p *pool) get() (net.Conn, bool, error) {
	select {
	case conn := <-p.free:
		return connWithDeadline(conn, 1*time.Second), false, nil
	default:
		conn, err := net.Dial("tcp", p.addr)
		return connWithDeadline(conn, 1*time.Second), true, err
	}
}

// put returns connection to pool
func (p *pool) put(conn net.Conn) {
	select {
	case p.free <- conn:
		return
	default:
		// queue is full, so this connection isn`t required anymore
		conn.Close()
	}
}

// Close closes all open connections
func (p *pool) close() {
	close(p.free)
	for conn := range p.free {
		conn.Close()
	}
}

// connWithDeadline sets deadline for connection
func connWithDeadline(c net.Conn, d time.Duration) net.Conn {
	if c == nil {
		return nil
	}

	c.SetDeadline(time.Now().Add(d))

	return c
}
