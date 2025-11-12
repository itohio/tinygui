package ui

import (
	"bytes"
	"io"
	"runtime"
)

// CommandStreamMux reads line-based commands from an io.Reader and dispatches
// them to registered handlers without extra allocations.
type CommandStreamMux struct {
	r         io.Reader
	c         map[string]func([]byte)
	buf, line []byte
}

// NewCommandStreamMux builds a mux for the provided reader and command map.
func NewCommandStreamMux(r io.Reader, commands map[string]func([]byte)) *CommandStreamMux {
	return &CommandStreamMux{
		r:    r,
		c:    commands,
		buf:  make([]byte, 1),      // Buffer for reading chunks
		line: make([]byte, 0, 128), // Buffer for accumulating a full line
	}
}

// Run reads commands from the reader and executes the mapped functions.
// It minimizes allocations by reusing buffers for line reading and parsing.
func (m *CommandStreamMux) Run() error {
	for {
		runtime.Gosched()
		// Read a chunk from the reader
		n, err := m.r.Read(m.buf)
		if err != nil {
			// Handle EOF or other errors
			if err == io.EOF {
				// Process any remaining data in line buffer
				m.processLine(m.line)
				return err
			}
			// For other errors, simply return
			return err
		}

		// Process the read chunk
		if n > 0 {
			// Append the chunk to our line buffer
			chunk := m.buf[:n]

			// Look for newlines in the chunk
			start := 0
			for i, b := range chunk {
				if b == '\n' || b == '\r' {
					// Add the segment before newline to our line buffer
					m.line = append(m.line, chunk[start:i]...)

					// Process the complete line
					m.processLine(m.line)

					// Reset line buffer by setting length to 0 but keeping capacity
					m.line = m.line[:0]

					start = i + 1
				}
			}

			// Add any remaining partial line data
			if start < len(chunk) {
				m.line = append(m.line, chunk[start:]...)
			}
		}
	}
}

// processLine parses a complete line and executes the appropriate command function
func (m *CommandStreamMux) processLine(line []byte) {
	// Skip empty lines
	if len(line) == 0 {
		return
	}

	// Find the command/argument separator (space)
	idx := bytes.IndexByte(line, ' ')

	var cmd string
	var arg []byte

	if idx == -1 {
		// No space found, the whole line is the command
		cmd = string(line)
	} else {
		// Split into command and argument
		cmd = string(line[:idx])
		if idx+1 < len(line) {
			arg = line[idx+1:]
		}
	}

	// Execute the command function if it exists
	if fn, exists := m.c[cmd]; exists {
		fn(arg)
	}
}
