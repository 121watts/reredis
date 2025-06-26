package wal

import "os"

type Writer struct {
	file *os.File
}

func NewWriter(filename string) (*Writer, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	
	return &Writer{
		file: file,
	}, nil
}

func (w *Writer) WriteCommand(cmd []string) error {
	encoded := EncodeArray(cmd)
	_, err := w.file.Write(encoded)
	if err != nil {
		return err
	}
	
	// Force write to disk
	return w.file.Sync()
}

func (w *Writer) Close() error {
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}
