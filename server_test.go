package logde

import (
	"fmt"
	"io"
	"net"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

var port int64 = 30000

func nextPort() int64 {
	return atomic.AddInt64(&port, 1)
}

type testClient struct {
	connection net.Conn
	addr       string
}

func (c *testClient) send(t *testing.T, request []byte, responseLength int) []byte {
	n, err := c.connection.Write(request)
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 4096)
	n, err = c.connection.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	return buf[:n]
}

func (c *testClient) assertRequest(t *testing.T, request []byte, expected []byte) {
	response := c.send(t, request, len(expected))

	if !reflect.DeepEqual(response, expected) {
		t.Fatalf("Expected: %s. Got: %s. Query: %s", expected, response, request)
	}
}

func testServer(t *testing.T) (*testClient, io.Closer) {
	addr := fmt.Sprintf(":%d", nextPort())
	server, err := New(DefaultConfig().WithAddr(addr))
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	go server.Run()

	connection, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	return &testClient{connection: connection, addr: addr}, server
}

func TestFailStartSamePort(t *testing.T) {
	addr := fmt.Sprintf(":%d", nextPort())
	_, err := New(DefaultConfig().WithAddr(addr))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	_, err = New(DefaultConfig().WithAddr(addr))
	if err == nil {
		t.Fatalf("Error expected")
	}
}

func TestSetGet(t *testing.T) {
	client, closer := testServer(t)
	defer closer.Close()

	client.assertRequest(t, []byte("SET foo 0 3\r\nbar\r\n"), resultOK)
	client.assertRequest(t, []byte("GET foo\r\n"), []byte("VALUES\r\n1\r\n3\r\nbar"))
}

func TestSetGetWithExpire(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	client, closer := testServer(t)
	defer closer.Close()

	client.assertRequest(t, []byte("SET foo bar 1\r\n"), resultOK)
	client.assertRequest(t, []byte("GET foo\r\n"), []byte("VALUES\r\n1\r\n3\r\nbar\r\n"))
	time.Sleep(1500 * time.Millisecond)
	client.assertRequest(t, []byte("GET foo\r\n"), []byte("NOT_FOUND\r\n"))
}

func TestHSetHGet(t *testing.T) {
	client, closer := testServer(t)
	defer closer.Close()

	client.assertRequest(t, []byte("HSET foo key1 0 3\r\nbar\r\n"), resultOK)
	client.assertRequest(t, []byte("HGET foo key1\r\n"), []byte("VALUES\r\n1\r\n3\r\nbar"))
	client.assertRequest(t, []byte("HGET foo key2\r\n"), []byte("NOT_FOUND\r\n"))
}

func TestHSetHGetWithExpire(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	client, closer := testServer(t)
	defer closer.Close()

	client.assertRequest(t, []byte("HSET foo key1 bar1\r\n"), resultOK)
	client.assertRequest(t, []byte("HSET foo key2 bar2 1\r\n"), resultOK)
	client.assertRequest(t, []byte("HSET foo key3 bar3 2\r\n"), resultOK)
	client.assertRequest(t, []byte("HGET foo key1\r\n"), []byte("DATA\r\n4 bar1\r\n"))
	client.assertRequest(t, []byte("HGET foo key2\r\n"), []byte("DATA\r\n4 bar2\r\n"))
	client.assertRequest(t, []byte("HGET foo key3\r\n"), []byte("DATA\r\n4 bar3\r\n"))

	time.Sleep(2 * time.Second)
	client.assertRequest(t, []byte("HGET foo key1\r\n"), []byte("DATA\r\n4 bar1\r\n"))
	client.assertRequest(t, []byte("HGET foo key2\r\n"), []byte("NOT_FOUND\r\n"))
	client.assertRequest(t, []byte("HGET foo key3\r\n"), []byte("DATA\r\n4 bar3\r\n"))

	time.Sleep(3 * time.Second)
	client.assertRequest(t, []byte("HGET foo key1\r\n"), []byte("DATA\r\n4 bar1\r\n"))
	client.assertRequest(t, []byte("HGET foo key2\r\n"), []byte("NOT_FOUND\r\n"))
	client.assertRequest(t, []byte("HGET foo key3\r\n"), []byte("NOT_FOUND\r\n"))
}

//func TestGetKeys(t *testing.T) {
//	client, closer := testServer(t)
//	defer closer.Close()
//
//	client.send(t, []byte("SET foo value\r\n"))
//	client.send(t, []byte("SET bar another_value\r\n"))
//
//	client.assertRequest(t, []byte("KEYS\r\n"), []byte("DATA\r\n2\r\n3\r\nfoo\r\n3\r\nbar\r\n"))
//}
