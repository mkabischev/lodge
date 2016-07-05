package testutil

import (
	"net"
	"testing"
	"time"
)

var (
	serverWaitMaxAttempts = 10
	serverDelay           = 5 * time.Millisecond
)

func NextListener(t *testing.T) (net.Listener, net.Conn) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}

	waitForListener(t, l)

	conn, err := net.Dial(l.Addr().Network(), l.Addr().String())
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	return l, conn
}

func waitForListener(t *testing.T, l net.Listener) {
	for i := 1; i <= serverWaitMaxAttempts; i++ {
		if _, err := net.Dial(l.Addr().Network(), l.Addr().String()); err == nil {
			return
		}

		time.Sleep(time.Duration(i) * serverDelay)
	}
}
