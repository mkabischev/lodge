package server

import (
	"errors"
	"strconv"
)

var (
	errBadFormat = errors.New("Bad format")
)

type command interface {
	arguments() int
	process(r *request, s Storage) ([]string, error)
}

type getCommand struct{}

func (c getCommand) arguments() int {
	return 1
}

func (c getCommand) process(r *request, s Storage) ([]string, error) {
	value, err := s.Get(r.arguments[0])
	if err != nil {
		return nil, err
	}

	return []string{value}, nil
}

type setCommand struct{}

func (c setCommand) arguments() int {
	return 3
}

func (c setCommand) process(r *request, s Storage) ([]string, error) {
	ttl, err := strconv.Atoi(r.arguments[1])
	if err != nil || ttl < 0 {
		return nil, errBadFormat
	}

	dataLength, err := strconv.Atoi(r.arguments[2])

	if err != nil || dataLength <= 0 {
		return nil, errBadFormat
	}

	data, err := r.data(dataLength)
	if err != nil {
		return nil, err
	}
	err = s.Set(r.arguments[0], string(data), int64(ttl))

	return nil, err
}

type hSetCommand struct{}

func (c hSetCommand) arguments() int {
	return 3
}

func (c hSetCommand) process(r *request, s Storage) ([]string, error) {
	dataLength, _ := strconv.Atoi(r.arguments[2])
	data, _ := r.data(dataLength)

	err := s.HSet(r.arguments[0], r.arguments[1], string(data))

	return nil, err
}

type hGetCommand struct{}

func (c hGetCommand) arguments() int {
	return 2
}

func (c hGetCommand) process(r *request, s Storage) ([]string, error) {
	value, err := s.HGet(r.arguments[0], r.arguments[1])
	if err != nil {
		return nil, err
	}

	return []string{value}, nil
}

type hGetAllCommand struct{}

func (c hGetAllCommand) arguments() int {
	return 1
}

func (c hGetAllCommand) process(r *request, s Storage) ([]string, error) {
	hash, err := s.HGetAll(r.arguments[0])
	if err != nil {
		return nil, err
	}

	values := make([]string, len(hash)*2)
	i := 0
	for key, value := range hash {
		values[i] = key
		values[i+1] = value
		i += 2
	}

	return values, nil
}

type keysCommand struct{}

func (c keysCommand) arguments() int {
	return 0
}

func (c keysCommand) process(r *request, s Storage) ([]string, error) {
	return s.Keys()
}

type deleteCommand struct{}

func (c deleteCommand) arguments() int {
	return 1
}

func (c deleteCommand) process(r *request, s Storage) ([]string, error) {
	err := s.Delete(r.arguments[0])

	return nil, err
}

type expireCommand struct{}

func (c expireCommand) arguments() int {
	return 2
}

func (c expireCommand) process(r *request, s Storage) ([]string, error) {
	ttl, err := strconv.Atoi(r.arguments[1])
	if err != nil || ttl < 0 {
		return nil, errBadFormat
	}

	return nil, s.Expire(r.arguments[0], int64(ttl))
}
