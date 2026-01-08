//go:build darwin

/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package runCmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
)

// osTTY runs a command in a pseudo-terminal and handles interactive prompts
// by waiting for specific strings and sending responses.
// If AsUser is specified, it will first login as that user before running the command.
// Returns all output from the command and any error that occurred.
func (r *Runner) osTTY(def Interactive) (string, error) {
	if r.logger != nil {
		if len(def.Command) > 0 {
			r.logger.Debugf(8360, "TTY executing command: %s (arguments redacted)", def.Command[0])
		}
	}
	// Set default timeout if not specified
	timeout := def.Timeout
	if timeout <= 0 {
		timeout = 60 // Default: 60 seconds
	}
	fmt.Printf("DEBUG: tty timeout %d\n", timeout)
	// Create context with timeout for hard kill
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	var actualActions []Action

	// If we need to login as a different user first
	if def.AsUser != nil {

		// Start with login command
		cmd = exec.CommandContext(ctx, "login", def.AsUser.Username)

		// Build the command string
		cmdString := strings.Join(def.Command, " ")

		// Prepend login actions to the action list
		var loginActions []Action

		if def.AsUser.RunAsRoot {
			// Run command as root using sudo -i (gets a root shell)
			// Disable echo to avoid matching our own commands
			loginActions = []Action{
				{
					WaitFor:  "Password:",
					Send:     def.AsUser.Password,
					Delay:    1000,
					DebugMsg: "Sending login password",
				},
				{
					WaitFor:  "",
					Send:     "echo __READY__",
					DebugMsg: "Sending ready marker",
				},
				{
					WaitFor:  "__READY__",
					Send:     "stty -echo",
					Delay:    100,
					DebugMsg: "Disabling terminal echo (1)",
				},
				{
					WaitFor:  "",
					Send:     "echo __ECHO_OFF__",
					DebugMsg: "Sending echo-off marker",
				},
				{
					WaitFor:  "__ECHO_OFF__",
					Send:     "sudo -K",
					DebugMsg: "Starting root shell (will prompt for password)",
					Delay:    500,
				},
				{
					WaitFor:  "",
					Send:     "sudo -i",
					DebugMsg: "Starting root shell (will prompt for password)",
				},
				{
					WaitFor:  "Password:",
					Send:     def.AsUser.Password,
					Delay:    1000,
					DebugMsg: "Sending sudo password",
				},
				{
					WaitFor:  "#",
					Send:     "stty -echo",
					Delay:    100,
					DebugMsg: "Disabling terminal echo in root shell",
				},
				{
					WaitFor:  "",
					Send:     "echo __ROOT__",
					DebugMsg: "Sending root marker",
				},
				{
					WaitFor:  "__ROOT__",
					Send:     cmdString + "; echo __DONE__",
					DebugMsg: "Sending command in root shell",
				},
			}
		} else {
			// Run command as the logged-in user
			loginActions = []Action{
				{
					WaitFor:  "Password:",
					Send:     def.AsUser.Password,
					Delay:    100,
					DebugMsg: "Sending login password",
				},
				{
					WaitFor:  "",
					Send:     "echo __READY__",
					DebugMsg: "Sending ready marker",
				},
				{
					WaitFor:  "__READY__",
					Send:     "stty -echo",
					Delay:    100,
					DebugMsg: "Disabling terminal echo",
				},
				{
					WaitFor:  "",
					Send:     "echo __ECHO_OFF__",
					DebugMsg: "Sending echo-off marker",
				},
				{
					WaitFor:  "__ECHO_OFF__",
					Send:     cmdString + "; echo __DONE__",
					DebugMsg: "Sending command",
				},
			}
		}

		// Combine login actions with any additional interactive actions
		actualActions = append(loginActions, def.Actions...)

		// Add appropriate exit commands based on whether we're in root shell or not
		if def.AsUser.RunAsRoot {
			// Exit root shell first, then user shell
			actualActions = append(actualActions,
				Action{
					WaitFor:  "__DONE__",
					Send:     "exit",
					Delay:    100,
					DebugMsg: "Exiting root shell",
				},
				Action{
					WaitFor:  "",
					Send:     "echo __ROOTEXITED__",
					DebugMsg: "Sending root-exited marker",
				},
				Action{
					WaitFor:  "__ROOTEXITED__",
					Send:     "exit",
					DebugMsg: "Exiting user shell",
				},
			)
		} else {
			// Just exit user shell
			actualActions = append(actualActions, Action{
				WaitFor:  "__DONE__",
				Send:     "exit",
				DebugMsg: "Exiting user shell",
			})
		}
	} else {

		// Normal execution without user switch
		cmd = exec.CommandContext(ctx, def.Command[0], def.Command[1:]...)
		actualActions = def.Actions
	}

	// Start the command with a pty so we can see interactive prompts
	// Note: pty.Start() manages its own SysProcAttr, so we don't override it
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to start command with pty: %w", err)
	}
	defer func(ptmx *os.File) {
		_ = ptmx.Close()
	}(ptmx)

	// Monitor context deadline and force kill if exceeded
	go func() {
		<-ctx.Done()
		// Only kill if it was actually a timeout (DeadlineExceeded), not a normal cancellation
		if errors.Is(ctx.Err(), context.DeadlineExceeded) && cmd.Process != nil {
			// Kill the process with SIGKILL and close PTY
			// The PTY will propagate the signal to child processes
			//fmt.Fprintf(os.Stderr, "*** TTY HARD TIMEOUT: Killing process %d with SIGKILL\n", cmd.Process.Pid) // TODO
			_ = cmd.Process.Signal(syscall.SIGKILL)
			_ = ptmx.Close() // Force close the PTY to cleanup children
		}
	}()

	// Buffer to accumulate all output
	var outputBuf bytes.Buffer

	// Handle the interactive prompts in a goroutine
	ioErrChan := make(chan error, 1)
	go func() {
		// Process each interaction in sequence
		for i, interaction := range actualActions {
			if interaction.DebugMsg != "" {
				//fmt.Fprintf(os.Stderr, "*** TTY DEBUG: %s\n", interaction.DebugMsg)
			}

			// If WaitFor is empty, just send immediately without waiting
			if interaction.WaitFor == "" {
				//fmt.Fprintf(os.Stderr, "*** TTY SEND (no wait): %s (length=%d)\n", interaction.Send, len(interaction.Send))
				_, err = ptmx.Write([]byte(interaction.Send + "\n"))
				if err != nil {
					ioErrChan <- fmt.Errorf("interaction %d failed to send: %w", i, err)
					return
				}
			} else {
				// Wait for prompt and send
				if err = ttyWaitAndSend(ptmx, &outputBuf, interaction.WaitFor, interaction.Send); err != nil {
					ioErrChan <- fmt.Errorf("interaction %d failed: %w", i, err)
					return
				}
			}

			// Apply delay if specified
			if interaction.Delay > 0 {
				//fmt.Fprintf(os.Stderr, "*** TTY DELAY: %dms\n", interaction.Delay)
				time.Sleep(time.Duration(interaction.Delay) * time.Millisecond)
			}
		}

		// Continue reading until EOF to capture any final output
		buf := make([]byte, 1024)
		var n int
		for {
			n, err = ptmx.Read(buf)
			if err != nil {
				if err == io.EOF {
					//fmt.Fprintf(os.Stderr, "*** TTY EOF: End of output stream\n")
					ioErrChan <- nil
					return
				}
				//fmt.Fprintf(os.Stderr, "*** TTY READ ERROR: %v\n", err)
				ioErrChan <- fmt.Errorf("error reading final output: %w", err)
				return
			}
			chunk := buf[:n]
			outputBuf.Write(chunk)
			//fmt.Fprintf(os.Stderr, "*** TTY DATA (%d bytes): %q\n", n, string(chunk))
		}
	}()

	// Wait for the command to complete in a separate goroutine to avoid deadlock
	cmdErrChan := make(chan error, 1)
	go func() {
		cmdErrChan <- cmd.Wait()
	}()

	// Wait for either the command or I/O operations to complete/fail
	var cmdErr, ioErr error
	cmdDone, ioDone := false, false

	for !cmdDone || !ioDone {
		select {
		case cmdErr = <-cmdErrChan:
			cmdDone = true
		case ioErr = <-ioErrChan:
			ioDone = true
			// If I/O fails (e.g., timeout), kill the process to prevent hanging
			if ioErr != nil && !cmdDone && cmd.Process != nil {
				//fmt.Fprintf(os.Stderr, "*** TTY I/O ERROR: Killing process %d with SIGKILL\n", cmd.Process.Pid)
				_ = cmd.Process.Signal(syscall.SIGKILL)
				_ = ptmx.Close() // Force close PTY to cleanup children
			}
		}
	}

	// Get all accumulated output
	output := outputBuf.String()

	// Check for I/O errors first (more specific than command errors)
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
	var matchBuf bytes.Buffer               // Separate buffer for pattern matching
	timeout := time.After(30 * time.Second) // Add timeout to prevent hanging

	//fmt.Fprintf(os.Stderr, "*** TTY WAIT: Waiting for '%s'\n", waitFor)

	for {
		select {
		case <-timeout:
			//fmt.Fprintf(os.Stderr, "*** TTY TIMEOUT: Never received '%s'. Buffer contents: %q\n", waitFor, matchBuf.String())
			return fmt.Errorf("timeout waiting for '%s'", waitFor)
		default:
			// Set a read deadline to allow checking timeout
			_ = ptmx.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

			n, err := ptmx.Read(buf)
			if err != nil {
				if os.IsTimeout(err) {
					continue // Just a timeout, keep trying
				}
				if err == io.EOF {
					return fmt.Errorf("unexpected EOF while waiting for '%s'", waitFor)
				}
				return fmt.Errorf("error reading from pty: %w", err)
			}

			// Clear the read deadline
			_ = ptmx.SetReadDeadline(time.Time{})

			// Append to both buffers
			chunk := buf[:n]
			outputBuf.Write(chunk) // Accumulate all output for final return
			matchBuf.Write(chunk)  // Use for pattern matching this call only

			// Debug: show received data
			//fmt.Fprintf(os.Stderr, "*** TTY RECV (%d bytes): %q\n", n, string(chunk))

			// Check if we've received the expected prompt in NEW data only
			if strings.Contains(matchBuf.String(), waitFor) {
				//fmt.Printf("\n\n--- HIT\n%s\n---\n\n", matchBuf.String())
				//fmt.Fprintf(os.Stderr, "*** TTY FOUND: Found '%s' in output\n", waitFor)
				time.Sleep(250 * time.Millisecond)
				_, err = ptmx.Write([]byte(sendValue + "\n"))
				if err != nil {
					return fmt.Errorf("error writing response: %w", err)
				}

				//fmt.Fprintf(os.Stderr, "*** TTY SENT: %s (length=%d)\n", sendValue, len(sendValue))

				// matchBuf is local and will be discarded, ensuring next wait only sees NEW data
				return nil
			}
		}
	}
}
