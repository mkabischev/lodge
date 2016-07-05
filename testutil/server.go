package testutil

import (
	"net"
	"sync/atomic"
	"testing"
	"time"
)

var port int64 = 30000

var serverWaitMaxAttempts = 10
var serverDelay = 5 * time.Millisecond

func NextPort() int64 {
	return atomic.AddInt64(&port, 1)
}

func WaitForAddr(t *testing.T, addr string) {
	for i := 1; i <= serverWaitMaxAttempts; i++ {
		if _, err := net.Dial("tcp", addr); err == nil {
			return
		}

		time.Sleep(time.Duration(i) * serverDelay)
	}
}
