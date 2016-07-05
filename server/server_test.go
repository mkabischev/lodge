package server

import (
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
	l, conn := testutil.NextListener(t)

	server := New(NewMemory())
	go server.Serve(l)


	return &testClient{connection: conn}, server
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

	client.assertRequest(t, []byte("SET foo 1 3\r\nbar\r\n"), resultOK)
	client.assertRequest(t, []byte("GET foo\r\n"), []byte("VALUES\r\n1\r\n3\r\nbar"))
	time.Sleep(2 * time.Second)
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

	client.assertRequest(t, []byte("HSET foo key1 0 4\r\nbar1\r\n"), resultOK)
	client.assertRequest(t, []byte("HSET foo key2 1 4\r\nbar2\r\n"), resultOK)
	client.assertRequest(t, []byte("HSET foo key3 2 4\r\nbar3\r\n"), resultOK)
	client.assertRequest(t, []byte("HGET foo key1\r\n"), []byte("VALUES\r\n1\r\n4\r\nbar1"))
	client.assertRequest(t, []byte("HGET foo key2\r\n"), []byte("VALUES\r\n1\r\n4\r\nbar2"))
	client.assertRequest(t, []byte("HGET foo key3\r\n"), []byte("VALUES\r\n1\r\n4\r\nbar3"))

	time.Sleep(2 * time.Second)
	client.assertRequest(t, []byte("HGET foo key1\r\n"), []byte("VALUES\r\n1\r\n4\r\nbar1"))
	client.assertRequest(t, []byte("HGET foo key2\r\n"), []byte("NOT_FOUND\r\n"))
	client.assertRequest(t, []byte("HGET foo key3\r\n"), []byte("VALUES\r\n1\r\n4\r\nbar3"))

	time.Sleep(3 * time.Second)
	client.assertRequest(t, []byte("HGET foo key1\r\n"), []byte("VALUES\r\n1\r\n4\r\nbar1"))
	client.assertRequest(t, []byte("HGET foo key2\r\n"), []byte("NOT_FOUND\r\n"))
	client.assertRequest(t, []byte("HGET foo key3\r\n"), []byte("NOT_FOUND\r\n"))
}
