package utils

import (
	"bufio"
	"fmt"
	"os"
)

// OutputWriter ENCAPSULATES bufio.Writer WITH ERROR HANDLING.
type OutputWriter struct {
	w *bufio.Writer
}

// NewOutputWriter CREATES A NEW OutputWriter
func NewOutputWriter() *OutputWriter {
	return &OutputWriter{
		w: bufio.NewWriter(os.Stdout),
	}
}

// Println WRITES A LINE OF TEXT AND HANDLES ERRORS.
func (ow *OutputWriter) Println(text string) bool {
	_, err := fmt.Fprintln(ow.w, text)

	if err != nil {
		return false
	}

	return ow.Flush()
}

// Printf WRITES FORMATTED TEXT AND HANDLES ERRORS.
func (ow *OutputWriter) Printf(format string, args ...interface{}) bool {
	_, err := fmt.Fprintf(ow.w, format, args...)

	if err != nil {
		return false
	}

	return ow.Flush()
}

// PrintNewLines WRITES EMPTY LINES.
func (ow *OutputWriter) PrintNewLines(count int) bool {
	for i := 0; i < count; i++ {
		if !ow.Println("") {
			return false
		}
	}

	return true
}

// Write WRITES DIRECTLY TO THE UNDERLYING WRITER.
func (ow *OutputWriter) Write(p []byte) (n int, err error) {
	return ow.w.Write(p)
}

// Flush EMPTIES THE BUFFER AND HANDLE ERRORS.
func (ow *OutputWriter) Flush() bool {
	return ow.w.Flush() == nil
}
