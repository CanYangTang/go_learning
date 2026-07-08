package files

import (
	"bufio"
	"os"
)

// WriteAll writes the entire content to the file.
func WriteAll(path string, content string) error {
	// TODO: implement using os.WriteFile
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return err
	}
	return nil
}

// WriteLines writes each line to the file, appending a newline after each line.
func WriteLines(path string, lines []string) error {
	// TODO: implement using bufio.Writer
	// 1. os.Create the file
	// 2. defer file.Close()
	// 3. bufio.NewWriter(file)
	// 4. loop lines and writer.WriteString(line + "\n")
	// 5. writer.Flush()
	// 6. return nil
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for _, line := range lines {
		writer.WriteString(line);
		writer.WriteString("\n")
	}
	writer.Flush()
	return nil
}
