package wal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Reader struct {
	scanner *bufio.Scanner
	file    *os.File
}

type Entry struct {
	Command []string
}

func (r *Reader) NewReader(filename string) (*Reader, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)

	return &Reader{
		file:    file,
		scanner: scanner,
	}, nil
}

func (r *Reader) ReadEntry() (*Entry, error) {
	command, err := r.parseArray()
	if err != nil {
		return nil, err
	}

	return &Entry{Command: command}, nil
}

func (r *Reader) parseArray() ([]string, error) {
	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return nil, err
		}

		return nil, io.EOF
	}

	line := r.scanner.Text()

	if !strings.HasPrefix(line, "*") {
		return nil, fmt.Errorf("expected array header, got: %s", line)
	}

	countStr := line[1:]
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return nil, fmt.Errorf("invalid array count: %s", countStr)
	}

	result := make([]string, count)
	for i := 0; i < count; i++ {
		bulkStr, err := r.parseBulkString()
		if err != nil {
			return nil, err
		}

		result[i] = bulkStr
	}

	return result, nil
}

func (r *Reader) parseBulkString() (string, error) {
	if !r.scanner.Scan() {
		return "", fmt.Errorf("unexpecteed EOF reading bulk string header")
	}

	line := r.scanner.Text()
	if !strings.HasPrefix(line, "$") {
		return "", fmt.Errorf("expected bulk string header, got: %s", line)
	}

	lengthStr := line[1:]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", fmt.Errorf("invalid bulk string length, got: %s", lengthStr)
	}

	if !r.scanner.Scan() {
		return "", fmt.Errorf("unexpected EOF reading bulk string data")
	}

	data := r.scanner.Text()
	if len(data) != length {
		return "", fmt.Errorf("bulks string length mismatch: expected %d, got %d", length, len(data))
	}

	return data, nil
}
