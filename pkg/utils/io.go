// Package utils provides utility functions for file I/O operations
package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ReadInput reads data from a file or STDIN
func ReadInput(path string) (io.Reader, error) {
	if path == "" || path == "-" {
		// Read from STDIN
		return os.Stdin, nil
	}

	// Read from file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open input file: %w", err)
	}

	return file, nil
}

// WriteFile writes data to a file atomically using a temporary file
func WriteFile(path string, data []byte, overwrite bool) error {
	// Check if file exists and overwrite is not allowed
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("file already exists: %s (use --overwrite to replace)", path)
		}
	}

	// Create parent directory if it doesn't exist
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Write to temporary file first
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Rename temporary file to final path (atomic on most systems)
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath) // Clean up on error
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ReadFileContent reads the entire content of a file
func ReadFileContent(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}
