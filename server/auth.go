package server

// Copyright: https://github.com/mindreframer/golang-stuff/blob/master/github.com/bitly/google_auth_proxy/htpasswd.go

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/csv"
	"io"
	"log"
	"os"
)

// lookup passwords in a htpasswd file
// The entries must have been created with -s for SHA encryption

type UserList struct {
	users map[string]string
}

func NewUserList(path string) (*UserList, error) {
	log.Printf("using htpasswd file %s", path)
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return NewUserListFromReader(r)
}

func NewUserListFromReader(file io.Reader) (*UserList, error) {
	csv_reader := csv.NewReader(file)
	csv_reader.Comma = ':'
	csv_reader.Comment = '#'
	csv_reader.TrimLeadingSpace = true

	records, err := csv_reader.ReadAll()
	if err != nil {
		return nil, err
	}
	h := &UserList{users: make(map[string]string)}
	for _, record := range records {
		h.users[record[0]] = record[1]
	}
	return h, nil
}

func (h *UserList) Validate(user string, password string) bool {
	realPassword, exists := h.users[user]
	if !exists {
		return false
	}
	if realPassword[:5] == "{SHA}" {
		d := sha1.New()
		d.Write([]byte(password))
		if realPassword[5:] == base64.StdEncoding.EncodeToString(d.Sum(nil)) {
			return true
		}
	} else {
		log.Printf("Invalid htpasswd entry for %s. Must be a SHA entry.", user)
	}
	return false
}
