//go:build linux

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

// TTY runs a command in a pseudo-terminal and handles interactive prompts
// by waiting for specific strings and sending responses.
// If AsUser is specified, it will first login as that user before running the command.
// Returns all output from the command and any error that occurred.
func osTTY(def Interactive) (string, error) {
	var cmd *exec.Cmd
	var actualActions []Action

	// If we need to login as a different user first
	if def.AsUser != nil {

		// Start with login command
		cmd = exec.Command("login", def.AsUser.Username)

		// Build the command string
		cmdString := strings.Join(def.Command, " ")

		// Prepend login actions to the action list
		var loginActions []Action

		if def.AsUser.RunAsRoot {
			// Run command as root using sudo
			loginActions = []Action{
				{
					WaitFor:  "Password:",
					Send:     def.AsUser.Password + "\necho __SHELL_READY__",
					DebugMsg: "Sending password and marker command",
				},
				{
					WaitFor:  "__SHELL_READY__",
					Send:     "sudo -S " + cmdString,
					DebugMsg: "Sending sudo command",
				},
				{
					WaitFor:  "Password:",
					Send:     def.AsUser.Password + "\necho __CMD_DONE__",
					DebugMsg: "Sending sudo password and completion marker",
				},
			}
		} else {
			// Run command as the logged-in user
			loginActions = []Action{
				{
					WaitFor:  "Password:",
					Send:     def.AsUser.Password + "\necho __SHELL_READY__",
					DebugMsg: "Sending password and marker command",
				},
				{
					WaitFor:  "__SHELL_READY__",
					Send:     cmdString + "\necho __CMD_DONE__",
					DebugMsg: "Sending command after shell ready",
				},
			}
		}

		// Combine login actions with original actions
		actualActions = append(loginActions, def.Actions...)

		// ALWAYS add exit command at the end, even if no interactions
		// Wait for command completion marker, then exit
		actualActions = append(actualActions, Action{
			WaitFor:  "__CMD_DONE__",
			Send:     "exit",
			DebugMsg: "Logging out",
		})
	} else {

		// Normal execution without user switch
		cmd = exec.Command(def.Command[0], def.Command[1:]...)
		actualActions = def.Actions
	}

	// Start the command with a pty so we can see interactive prompts
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to start command with pty: %w", err)
	}
	defer ptmx.Close()

	// Buffer to accumulate all output
	var outputBuf bytes.Buffer

	// Handle the interactive prompts in a goroutine
	ioErrChan := make(chan error, 1)
	go func() {
		// Process each interaction in sequence
		for i, interaction := range actualActions {
			if interaction.DebugMsg != "" {
				// Uncomment for debugging
				// fmt.Printf("*** TTY DEBUG: %s\n", interaction.DebugMsg)
			}

			if err := ttyWaitAndSend(ptmx, &outputBuf, interaction.WaitFor, interaction.Send); err != nil {
				ioErrChan <- fmt.Errorf("interaction %d failed: %w", i, err)
				return
			}
		}

		// Continue reading until EOF to capture any final output
		buf := make([]byte, 1024)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				if err == io.EOF {
					ioErrChan <- nil
					return
				}
				ioErrChan <- fmt.Errorf("error reading final output: %w", err)
				return
			}
			outputBuf.Write(buf[:n])
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
			// If I/O fails (e.g., timeout), kill the command to prevent hanging
			if ioErr != nil && !cmdDone {
				cmd.Process.Kill()
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
	timeout := time.After(30 * time.Second) // Add timeout to prevent hanging

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for '%s'", waitFor)
		default:
			// Set a read deadline to allow checking timeout
			ptmx.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

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
			ptmx.SetReadDeadline(time.Time{})

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
				// Don't clear the buffer here - keep accumulating
				// outputBuf.Reset()
				return nil
			}
		}
	}
}
