package session

import (
	"bytes"
	"os"
	"strings"
)

func readLastNLines(filePath string, n int, maxBytes int64) ([]string, error) {
	// Validate inputs
	if n <= 0 || maxBytes <= 0 {
		return []string{}, nil
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	// Close the file when done
	defer file.Close()

	// Get the file size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := fileInfo.Size()
	if fileSize == 0 {
		return []string{}, nil
	}

	// Define our variables
	var lines []string                      // To store the last N lines
	var readRange = min(fileSize, maxBytes) // Number of bytes to read
	var offset = fileSize - readRange       // Offset to start reading from

	// Read the file from the calculated offset
	var buffer = make([]byte, readRange)
	_, err = file.ReadAt(buffer, offset)
	if err != nil {
		return nil, err
	}

	// Ensure valid UTF-8 and split into lines
	buffer = bytes.ToValidUTF8(buffer, []byte{'?'})

	// Convert buffer to string
	var logs = string(buffer)

	logs = strings.ReplaceAll(logs, "\r\n", "\n") // Handle Windows line endings into Unix
	logs = strings.ReplaceAll(logs, "\r", "\n")   // Handle old Mac line endings into Unix

	// Split the logs into lines
	lines = strings.Split(logs, "\n")

	// Clean up empty lines at end
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// Handle partial first line
	if offset > 0 && len(lines) > 0 && logs[0] != '\n' {
		lines = lines[1:] // Remove potentially partial first line
	}

	// If we have more lines than needed, trim the slice
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	return lines, nil
}
