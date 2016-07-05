package server

import (
	"fmt"
	"io"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/mkabischev/lodge/testutil"
)

type testClient struct {
	connection net.Conn
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
	addr := fmt.Sprintf(":%d", testutil.NextPort())
	server, err := New(DefaultConfig().WithAddr(addr))
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	go server.Run()

	testutil.WaitForAddr(t, addr)

	connection, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	return &testClient{connection: connection}, server
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
