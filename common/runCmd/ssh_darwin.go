//go:build darwin

/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
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
func (r *Runner) SSH(user *UserLogin, cmdAndArgs ...string) (string, error) {
	if user == nil {
		return "", fmt.Errorf("user credentials required")
	}
	if len(cmdAndArgs) == 0 {
		return "", fmt.Errorf("no command specified")
	}

	if r.logger != nil {
		r.logger.Debugf(8300, "SSH command requested: %s (arguments redacted) (user: %s, runAsRoot: %v)", cmdAndArgs[0], user.Username, user.RunAsRoot)
	}

	// Check if SSH is running and start if needed
	sshWasStarted, err := r.ensureSSHRunning()
	if err != nil {
		if r.logger != nil {
			r.logger.Errorf(8301, "failed to ensure SSH is running: %v", err)
		}
		return "", fmt.Errorf("failed to ensure SSH is running: %w", err)
	}

	// Ensure we stop SSH if we started it
	defer func() {
		if sshWasStarted {
			if r.logger != nil {
				r.logger.Debugf(8302, "stopping SSH service that we started")
			}
			err := r.stopSSH()
			if err != nil {
				if r.logger != nil {
					r.logger.Warningf(8303, "error stopping SSH: %v", err)
				}
			} else {
				if r.logger != nil {
					r.logger.Debugf(8304, "SSH service stopped successfully")
				}
			}
		}
	}()

	// Give SSH a moment to fully start
	if sshWasStarted {
		if r.logger != nil {
			r.logger.Debugf(8305, "waiting 500ms for SSH to fully initialize")
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Build command string
	cmdString := strings.Join(cmdAndArgs, " ")

	// Execute via SSH
	if r.logger != nil {
		r.logger.Debugf(8306, "attempting SSH connection to %s as user %s", sshHost, user.Username)
	}
	output, err := r.executeSSH(user, cmdString)
	if err != nil {
		if r.logger != nil {
			r.logger.Warningf(8307, "SSH command execution failed: %v", err)
		}
		return output, err
	}

	if r.logger != nil {
		r.logger.Debugf(8308, "SSH command execution succeeded")
	}

	return output, nil
}

// ensureSSHRunning checks if SSH is running and starts it if not.
// Returns true if SSH was started by this function.
func (r *Runner) ensureSSHRunning() (bool, error) {
	// Check if SSH is already running
	if r.logger != nil {
		r.logger.Debugf(8310, "checking SSH service status via systemsetup -getremotelogin")
	}
	output, err := r.Combined("systemsetup", "-getremotelogin")
	if err != nil {
		if r.logger != nil {
			r.logger.Errorf(8311, "error checking SSH status: %v, output: %s", err, output)
		}
		return false, fmt.Errorf("failed to check SSH status: %w", err)
	}

	if strings.Contains(output, "Remote Login: On") {
		// SSH is already running
		if r.logger != nil {
			r.logger.Debugf(8312, "SSH already running (Remote Login: On)")
		}
		return false, nil
	}

	// SSH not running, start it
	if r.logger != nil {
		r.logger.Infof(8313, "SSH not running (Remote Login: Off), starting with: launchctl load -w %s", sshPlistPath)
	}
	output, err = r.Combined("launchctl", "load", "-w", sshPlistPath)
	if err != nil {
		if r.logger != nil {
			r.logger.Errorf(8314, "failed to start SSH: %v, output: %s", err, output)
		}
		return false, fmt.Errorf("failed to start SSH (launchctl load returned error): %w", err)
	}

	if r.logger != nil {
		r.logger.Debugf(8315, "SSH start command succeeded, output: %s", output)
	}

	// Verify SSH actually started
	if r.logger != nil {
		r.logger.Debugf(8316, "verifying SSH service started")
	}
	output, err = r.Combined("systemsetup", "-getremotelogin")
	if err != nil {
		if r.logger != nil {
			r.logger.Errorf(8317, "SSH verification check failed: %v, output: %s", err, output)
		}
		return false, fmt.Errorf("failed to verify SSH status after starting: %w", err)
	}

	if !strings.Contains(output, "Remote Login: On") {
		if r.logger != nil {
			r.logger.Warningf(8318, "SSH verification failed: Remote Login not On, output: %s", output)
		}
		return false, fmt.Errorf("SSH start command succeeded but Remote Login is not On")
	}

	if r.logger != nil {
		r.logger.Infof(8319, "SSH verified running (Remote Login: On)")
	}
	return true, nil
}

// stopSSH stops the SSH service
func (r *Runner) stopSSH() error {
	if r.logger != nil {
		r.logger.Debugf(8320, "stopping SSH service with: launchctl unload -w %s", sshPlistPath)
	}
	output, err := r.Combined("launchctl", "unload", "-w", sshPlistPath)
	if err != nil && r.logger != nil {
		r.logger.Debugf(8321, "SSH stop command returned error: %v, output: %s", err, output)
	} else if r.logger != nil {
		r.logger.Debugf(8322, "SSH stop command succeeded")
	}
	return err
}

// executeSSH connects via SSH and executes a command
func (r *Runner) executeSSH(user *UserLogin, command string) (string, error) {
	if r.logger != nil {
		r.logger.Debugf(8325, "configuring SSH client for connection to %s", sshHost)
	}
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
	if r.logger != nil {
		r.logger.Debugf(8326, "connecting to SSH server at %s as user %s", sshHost, user.Username)
	}
	client, err := ssh.Dial("tcp", sshHost, config)
	if err != nil {
		if r.logger != nil {
			r.logger.Errorf(8327, "failed to connect to SSH: %v", err)
		}
		return "", fmt.Errorf("failed to connect to SSH: %w", err)
	}
	defer func(client *ssh.Client) {
		_ = client.Close()
	}(client)

	if r.logger != nil {
		r.logger.Debugf(8328, "SSH connection established, creating session")
	}

	// Create session
	session, err := client.NewSession()
	if err != nil {
		if r.logger != nil {
			r.logger.Errorf(8329, "failed to create SSH session: %v", err)
		}
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	// If running as root, we need a PTY for sudo password prompt
	if user.RunAsRoot {
		if r.logger != nil {
			r.logger.Debugf(8330, "executing command with sudo (command redacted)")
		}
		return r.executeWithSudo(session, user.Password, command)
	}

	// Simple execution without sudo
	if r.logger != nil {
		r.logger.Debugf(8331, "executing command as user %s (command redacted)", user.Username)
	}
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(command)
	output := stdout.String() + stderr.String()

	if err != nil {
		if r.logger != nil {
			r.logger.Debugf(8332, "command execution failed: %v, output: %s", err, output)
		}
		return output, fmt.Errorf("command failed: %w", err)
	}

	if r.logger != nil {
		r.logger.Debugf(8333, "command execution succeeded, output length: %d bytes", len(output))
	}

	return output, nil
}

// executeWithSudo runs a command via sudo with PTY for password prompt
func (r *Runner) executeWithSudo(session *ssh.Session, password, command string) (string, error) {
	if r.logger != nil {
		r.logger.Debugf(8335, "requesting PTY for sudo execution")
	}
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
