package tools

import (
	"fmt"
	"io"
	"os"
)

// Output provides consistent output formatting across all commands
type Output struct {
	writer io.Writer
}

// NewOutput creates a new Output formatter
func NewOutput(w io.Writer) *Output {
	if w == nil {
		w = os.Stdout
	}
	return &Output{writer: w}
}

// DefaultOutput returns an Output formatter writing to stdout
func DefaultOutput() *Output {
	return NewOutput(os.Stdout)
}

// Success prints a success message with checkmark
func (o *Output) Success(message string) {
	fmt.Fprintf(o.writer, "✓ %s\n", message)
}

// Error prints an error message with X
func (o *Output) Error(message string) {
	fmt.Fprintf(o.writer, "✗ %s\n", message)
}

// Info prints an informational message
func (o *Output) Info(message string) {
	fmt.Fprintf(o.writer, "%s\n", message)
}

// Progress prints a progress message (operation in progress)
func (o *Output) Progress(message string) {
	fmt.Fprintf(o.writer, "%s\n", message)
}

// Header prints a header/section title
func (o *Output) Header(message string) {
	fmt.Fprintf(o.writer, "\n%s\n\n", message)
}

// FilesList prints a list of files
func (o *Output) FilesList(title string, files []string) {
	fmt.Fprintf(o.writer, "%s:\n", title)
	for _, file := range files {
		fmt.Fprintf(o.writer, "  • %s\n", file)
	}
}

// NextSteps prints a numbered list of next steps
func (o *Output) NextSteps(steps []string) {
	fmt.Fprintln(o.writer, "\nNext steps:")
	for i, step := range steps {
		fmt.Fprintf(o.writer, "  %d. %s\n", i+1, step)
	}
}

// EmptyLine prints an empty line
func (o *Output) EmptyLine() {
	fmt.Fprintln(o.writer)
}
