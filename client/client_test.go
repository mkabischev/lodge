package client

import (
	"io"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/mkabischev/lodge/server"
	"github.com/mkabischev/lodge/testutil"
)

func testServer(t *testing.T) (*Client, io.Closer) {
	l, _ := testutil.NextListener(t)

	server := server.New(server.NewMemory())
	go server.Serve(l)

	return New(Config{addr: l.Addr().String()}), server
}

func TestGetNonExistingKey(t *testing.T) {
	client, closer := testServer(t)
	defer closer.Close()

	assertKeyNotFound(t, client, "foo")
}

func TestSetGet(t *testing.T) {
	client, closer := testServer(t)
	defer closer.Close()

	cases := []struct {
		key   string
		value string
	}{
		{"key1", "bar"},
		{"key2", "value with space"},
		{"key3", "com plex\nda\n\rta"},
	}

	for _, tc := range cases {
		if err := client.Set(tc.key, tc.value, 0); err != nil {
			t.Fatalf("Unexpectec error: %v", err)
		}

		assertKey(t, client, tc.key, tc.value)
	}
}

func TestHSetHGet(t *testing.T) {
	client, closer := testServer(t)
	defer closer.Close()

	cases := []struct {
		key   string
		field string
		value string
	}{
		{"key1", "field1", "bar"},
		{"key1", "field1", "value with space"},
		{"key1", "field2", "com plex\nda\n\rta"},
	}

	for _, tc := range cases {
		if err := client.HSet(tc.key, tc.field, tc.value, 0); err != nil {
			t.Fatalf("Unexpectec error: %v", err)
		}

		value, err := client.HGet(tc.key, tc.field)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if value != tc.value {
			t.Fatalf("Expected: %v. Got: %v", tc.value, value)
		}
	}
}

func TestHGetAll(t *testing.T) {
	client, closer := testServer(t)
	defer closer.Close()

	client.HSet("key", "field1", "value1", 0)
	client.HSet("key", "field2", "value2", 0)

	result, err := client.HGetAll("key")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
	}

	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("Expected: %v. Got: %v", expected, result)
	}
}

func TestUpdate(t *testing.T) {
	client, closer := testServer(t)
	defer closer.Close()

	client.Set("foo", "bar", 0)
	client.Set("foo", "xyz", 0)

	assertKey(t, client, "foo", "xyz")
}

func TestSetGetExpire(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	client, closer := testServer(t)
	defer closer.Close()

	if err := client.Set("foo", "bar", 1); err != nil {
		t.Fatalf("Unexpectec error: %v", err)
	}

	assertKeyExists(t, client, "foo")
	time.Sleep(2 * time.Second)
	assertKeyNotFound(t, client, "foo")
}

func TestKeys(t *testing.T) {
	client, closer := testServer(t)
	defer closer.Close()

	client.Set("foo", "bar", 0)
	client.Set("xyz", "qwerty", 0)

	keys, err := client.Keys()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []string{"foo", "xyz"}

	// server returns keys in random order, so for testing it`s required to sort them
	sort.Strings(expected)
	sort.Strings(keys)
	if !reflect.DeepEqual(expected, keys) {
		t.Fatalf("Expected: %v. Got: %v", expected, keys)
	}
}

func TestDelete(t *testing.T) {
	client, closer := testServer(t)
	defer closer.Close()

	client.Set("foo", "bar", 0)
	client.Delete("foo")

	assertKeyNotFound(t, client, "foo")
}

func assertKeyExists(t *testing.T, c *Client, key string) {
	_, err := c.Get(key)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func assertKey(t *testing.T, c *Client, key, expected string) {
	value, err := c.Get(key)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if value != expected {
		t.Fatalf("Expected: %v. Got: %v", expected, value)
	}
}

func assertKeyNotFound(t *testing.T, c *Client, key string) {
	_, err := c.Get(key)
	if err != ErrNotFound {
		t.Fatalf("Expected ErrNotFound. Got error: %v", err)
	}
}
