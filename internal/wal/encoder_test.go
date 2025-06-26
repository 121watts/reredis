package wal

import (
	"testing"
)

func TestEncodeBulkString(t *testing.T) {
	// Test case: "hello" should become $5\r\nhello\r\n
	result := EncodeBulkString("hello")
	expected := []byte("$5\r\nhello\r\n")
	
	if string(result) != string(expected) {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestEncodeArray(t *testing.T) {
	// Test case: SET command
	result := EncodeArray([]string{"SET", "mykey", "myvalue"})
	expected := "*3\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$7\r\nmyvalue\r\n"
	
	if string(result) != expected {
		t.Errorf("Expected %q, got %q", expected, string(result))
	}
}