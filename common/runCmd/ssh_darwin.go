//go:build darwin

/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package runCmd

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	sshPlistPath = "/System/Library/LaunchDaemons/ssh.plist"
	sshHost      = "127.0.0.1:22"
)

// SSH runs a command via SSH to localhost using the provided credentials.
// If RunAsRoot is true, it will use sudo to execute the command as root.
// SSH will be automatically started if not running and stopped afterward.
func SSH(user *UserLogin, cmdAndArgs ...string) (string, error) {
	if user == nil {
		return "", fmt.Errorf("user credentials required")
	}
	if len(cmdAndArgs) == 0 {
		return "", fmt.Errorf("no command specified")
	}

	// Check if SSH is running and start if needed
	sshWasStarted, err := ensureSSHRunning()
	if err != nil {
		return "", fmt.Errorf("failed to ensure SSH is running: %w", err)
	}

	// Ensure we stop SSH if we started it
	defer func() {
		if sshWasStarted {
			_ = stopSSH()
		}
	}()

	// Give SSH a moment to fully start
	if sshWasStarted {
		time.Sleep(500 * time.Millisecond)
	}

	// Build command string
	cmdString := strings.Join(cmdAndArgs, " ")

	// Execute via SSH
	output, err := executeSSH(user, cmdString)
	if err != nil {
		return output, err
	}

	return output, nil
}

// ensureSSHRunning checks if SSH is running and starts it if not.
// Returns true if SSH was started by this function.
func ensureSSHRunning() (bool, error) {
	// Check if SSH is already running
	_, err := Combined("launchctl", "list", "com.openssh.sshd")
	if err == nil {
		// SSH is already running
		return false, nil
	}

	// SSH not running, start it
	_, err = Combined("launchctl", "load", sshPlistPath)
	if err != nil {
		return false, fmt.Errorf("failed to start SSH: %w", err)
	}

	return true, nil
}

// stopSSH stops the SSH service
func stopSSH() error {
	_, err := Combined("launchctl", "unload", sshPlistPath)
	return err
}

// executeSSH connects via SSH and executes a command
func executeSSH(user *UserLogin, command string) (string, error) {
	// Configure SSH client
	config := &ssh.ClientConfig{
		User: user.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(user.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Safe for localhost
		Timeout:         10 * time.Second,
	}

	// Connect to SSH server
	client, err := ssh.Dial("tcp", sshHost, config)
	if err != nil {
		return "", fmt.Errorf("failed to connect to SSH: %w", err)
	}
	defer client.Close()

	// Create session
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// If running as root, we need a PTY for sudo password prompt
	if user.RunAsRoot {
		return executeWithSudo(session, user.Password, command)
	}

	// Simple execution without sudo
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(command)
	output := stdout.String() + stderr.String()

	if err != nil {
		return output, fmt.Errorf("command failed: %w", err)
	}

	return output, nil
}

// executeWithSudo runs a command via sudo with PTY for password prompt
func executeWithSudo(session *ssh.Session, password, command string) (string, error) {
	// Request a PTY for interactive sudo
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // Disable echo
		ssh.TTY_OP_ISPEED: 14400, // Input speed
		ssh.TTY_OP_OSPEED: 14400, // Output speed
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return "", fmt.Errorf("failed to request PTY: %w", err)
	}

	// Set up pipes for I/O
	stdin, err := session.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Buffer for output
	var outputBuf bytes.Buffer

	// Start the sudo command
	// Use sudo -S to read password from stdin, -p for custom prompt we can detect
	sudoCmd := fmt.Sprintf("sudo -S -p 'SUDO_PROMPT:' %s", command)

	if err := session.Start(sudoCmd); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Read output and handle sudo prompt
	go func() {
		buf := make([]byte, 1024)
		promptSent := false

		for {
			n, err := stdout.Read(buf)
			if err != nil {
				if err != io.EOF {
					// Log error but don't fail
				}
				return
			}

			chunk := string(buf[:n])
			outputBuf.WriteString(chunk)

			// Check for sudo password prompt
			if !promptSent && strings.Contains(outputBuf.String(), "SUDO_PROMPT:") {
				// Send password
				_, _ = stdin.Write([]byte(password + "\n"))
				promptSent = true
			}
		}
	}()

	// Wait for command to complete
	err = session.Wait()

	// Give a moment for output to be captured
	time.Sleep(100 * time.Millisecond)

	output := outputBuf.String()

	// Clean up the output - remove the sudo prompt
	output = strings.ReplaceAll(output, "SUDO_PROMPT:", "")

	if err != nil {
		return output, fmt.Errorf("command failed: %w", err)
	}

	return output, nil
}
