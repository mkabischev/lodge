package server

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/mkabischev/logde/ioutil"
)

type request struct {
	command   string
	arguments []string

	reader *bufio.Reader
}

func Parse(reader io.Reader) (*request, error) {
	r := &request{
		reader: bufio.NewReader(reader),
	}

	if err := r.parseHeader(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *request) parseHeader() error {
	header, _, err := r.reader.ReadLine()
	if err != nil {
		return err
	}

	fields := strings.Fields(string(header))
	if len(fields) == 0 {
		return fmt.Errorf("Empty request")
	}

	r.command = fields[0]
	r.arguments = fields[1:]

	return nil
}

func (r *request) data(n int) ([]byte, error) {
	return ioutil.Read(r.reader, n)
}
