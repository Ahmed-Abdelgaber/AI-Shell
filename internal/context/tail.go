package context

import (
	"bufio"
	"bytes"
	"io"
	"os"
)

func readLastNLines(filePath string, n int) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Seek to the end of the file
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := fileInfo.Size()

	// Read in chunks from the end
	const chunkSize = int64(4096) // Adjust as needed
	var lines []string
	var currentPos int64 = fileSize

	for {
		readSize := chunkSize
		if currentPos-chunkSize < 0 {
			readSize = currentPos
		}
		currentPos -= readSize

		_, err := file.Seek(currentPos, io.SeekStart)
		if err != nil {
			return nil, err
		}

		buffer := make([]byte, readSize)
		_, err = file.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, err
		}

		scanner := bufio.NewScanner(io.MultiReader(bytes.NewReader(buffer), file))
		for scanner.Scan() {
			lines = append([]string{scanner.Text()}, lines...) // Prepend to maintain order
			if len(lines) > n {
				lines = lines[:n] // Keep only the last N lines
			}
		}

		if currentPos == 0 || len(lines) >= n {
			break
		}
	}

	return lines, nil
}
