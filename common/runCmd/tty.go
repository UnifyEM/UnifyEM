/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package runCmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/creack/pty"
)

// Interactive represents a command and series of interactions
type Interactive struct {
	Command []string
	Actions []Action
}

// Action represents a single prompt/response interaction in an interactive TTY session
type Action struct {
	WaitFor  string // The string to wait for in the output
	Send     string // The value to send when the prompt is detected
	DebugMsg string // Optional debug message to print when this interaction occurs
}

// TTY runs a command in a pseudo-terminal and handles interactive prompts
// by waiting for specific strings and sending responses.
// Returns all output from the command and any error that occurred.
func TTY(def Interactive) (string, error) {

	// Create the command
	cmd := exec.Command(def.Command[0], def.Command[1:]...)

	// Start the command with a pty so we can see interactive prompts
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to start command with pty: %w", err)
	}
	defer ptmx.Close()

	// Buffer to accumulate all output
	var outputBuf bytes.Buffer

	// Handle the interactive prompts in a goroutine
	errChan := make(chan error, 1)
	go func() {
		// Process each interaction in sequence
		for i, interaction := range def.Actions {
			//if interaction.DebugMsg != "" {
			//	fmt.Printf("*** TTY DEBUG: %s\n", interaction.DebugMsg)
			//}

			if err := ttyWaitAndSend(ptmx, &outputBuf, interaction.WaitFor, interaction.Send); err != nil {
				errChan <- fmt.Errorf("interaction %d failed: %w", i, err)
				return
			}
		}

		// Continue reading until EOF to capture any final output
		buf := make([]byte, 1024)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				if err == io.EOF {
					errChan <- nil
					return
				}
				errChan <- fmt.Errorf("error reading final output: %w", err)
				return
			}
			outputBuf.Write(buf[:n])
		}
	}()

	// Wait for the command to complete
	cmdErr := cmd.Wait()
	ioErr := <-errChan

	// Get all accumulated output
	output := outputBuf.String()

	// Check for I/O errors first
	if ioErr != nil {
		return output, ioErr
	}

	// Return command error if present
	if cmdErr != nil {
		return output, fmt.Errorf("command failed: %w", cmdErr)
	}

	return output, nil
}

// ttyWaitAndSend waits for a specific prompt string in the PTY output and sends a response
func ttyWaitAndSend(ptmx *os.File, outputBuf *bytes.Buffer, waitFor string, sendValue string) error {
	buf := make([]byte, 1024)
	for {
		n, err := ptmx.Read(buf)
		if err != nil {
			if err == io.EOF {
				return fmt.Errorf("unexpected EOF while waiting for '%s'", waitFor)
			}
			return fmt.Errorf("error reading from pty: %w", err)
		}

		// Append to output buffer
		chunk := buf[:n]
		outputBuf.Write(chunk)

		// Check if we've received the expected prompt
		if strings.Contains(outputBuf.String(), waitFor) {
			time.Sleep(250 * time.Millisecond)
			_, err = ptmx.Write([]byte(sendValue + "\n"))
			if err != nil {
				return fmt.Errorf("error writing response: %w", err)
			}
			// Clear the buffer after successful match and response
			outputBuf.Reset()
			return nil
		}
	}
}
