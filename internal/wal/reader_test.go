package wal

import (
	"io"
	"os"
	"testing"
)

func TestWriteReadRoundTrip(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "wal_test_*.wal")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Test commands to write and read back
	testCommands := [][]string{
		{"SET", "key1", "value1"},
		{"SET", "key2", "value with spaces"},
		{"DEL", "key1"},
		{"SET", "key3", ""},
		{"DEL", "key2"},
		{"SET", "long_key", "a very long value with multiple words and special chars!@#$%"},
	}

	// Write commands
	writer, err := NewWriter(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}

	for _, cmd := range testCommands {
		err := writer.WriteCommand(cmd)
		if err != nil {
			t.Fatalf("Failed to write command %v: %v", cmd, err)
		}
	}
	writer.Close()

	// Read commands back
	reader, err := NewReader(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	readCommands := [][]string{}
	for {
		entry, err := reader.ReadEntry()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to read entry: %v", err)
		}
		readCommands = append(readCommands, entry.Command)
	}

	// Compare
	if len(readCommands) != len(testCommands) {
		t.Fatalf("Command count mismatch: expected %d, got %d", len(testCommands), len(readCommands))
	}

	for i, expectedCmd := range testCommands {
		actualCmd := readCommands[i]
		if len(actualCmd) != len(expectedCmd) {
			t.Errorf("Command %d length mismatch: expected %d, got %d", i, len(expectedCmd), len(actualCmd))
			continue
		}

		for j, expectedPart := range expectedCmd {
			if actualCmd[j] != expectedPart {
				t.Errorf("Command %d part %d mismatch: expected %q, got %q", i, j, expectedPart, actualCmd[j])
			}
		}
	}
}