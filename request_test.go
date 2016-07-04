package logde

import (
	"bytes"
	"testing"
	"reflect"
)

func TestSuccessParse(t *testing.T) {
	cases := []struct {
		requestString string
		command       string
		args          []string
	}{
		{"SET foo 0 3\r\nbar\r\n", "SET", []string{"foo", "0", "3"}},
		{"GET foo\r\n", "GET", []string{"foo"}},
		{"HSET foo bar 100 5\r\nhello\r\n", "HSET", []string{"foo", "bar", "100", "5"}},
	}

	for _, tc := range cases {
		r, err := Parse(bytes.NewBufferString(tc.requestString))

		if err != nil {
			t.Fatalf(`Unexpected error "%v" for "%s"`, err, tc.requestString)
		}

		if r.command != tc.command {
			t.Fatalf("Expected: %v. Got: %v", tc.command, r.command)
		}

		if !reflect.DeepEqual(r.arguments, tc.args) {
			t.Fatalf("Expected: %v. Got: %v", tc.args, r.arguments)
		}
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Parse(bytes.NewBufferString("LPUSH foo 1 2 3 4"))
	}
}

//func args(args ...string) [][]byte {
//	res := make([][]byte, len(args))
//
//	for i, arg := range args {
//		res[i] = []byte(arg)
//	}
//
//	return res
//}
