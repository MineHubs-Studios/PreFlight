package utils

import (
	"bufio"
	"fmt"
	"os"
)

// OutputWriter encapsulates bufio.Writer with error handling.
type OutputWriter struct {
	w *bufio.Writer
}

// NewOutputWriter creates a new OutputWriter
func NewOutputWriter() *OutputWriter {
	return &OutputWriter{
		w: bufio.NewWriter(os.Stdout),
	}
}

// Println writes a line of text and handles errors.
func (ow *OutputWriter) Println(text string) bool {
	_, err := fmt.Fprintln(ow.w, text)

	if err != nil {
		return false
	}

	return ow.Flush()
}

// Printf writes formatted text and handles errors.
func (ow *OutputWriter) Printf(format string, args ...interface{}) bool {
	_, err := fmt.Fprintf(ow.w, format, args...)

	if err != nil {
		return false
	}

	return ow.Flush()
}

// PrintNewLines writes empty lines.
func (ow *OutputWriter) PrintNewLines(count int) bool {
	for i := 0; i < count; i++ {
		if !ow.Println("") {
			return false
		}
	}

	return true
}

// Write writes directly to the underlying writer.
func (ow *OutputWriter) Write(p []byte) (n int, err error) {
	return ow.w.Write(p)
}

// Flush empties the buffer and handle errors.
func (ow *OutputWriter) Flush() bool {
	return ow.w.Flush() == nil
}
