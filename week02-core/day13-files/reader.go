package files

import (
	"bufio"
	"os"
)

// ReadAll reads the entire file content and returns it as a string.
func ReadAll(path string) (string, error) {
	// TODO: implement using os.ReadFile
	file, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(file), nil
}

// ReadLines reads the file line by line and returns each line as a string slice.
func ReadLines(path string) ([]string, error) {
	// TODO: implement using bufio.Scanner
	// 1. os.Open the file
	// 2. defer file.Close()
	// 3. bufio.NewScanner(file)
	// 4. loop scanner.Scan() and collect scanner.Text()
	// 5. return lines and scanner.Err()
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	lines := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
