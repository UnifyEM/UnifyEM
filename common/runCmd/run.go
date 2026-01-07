/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package runCmd

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"

	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

const (
	RunCombined int = iota
	RunSeparate
	RunStdout
	RunStderr
)

// Runner provides command execution with optional logging
type Runner struct {
	logger interfaces.Logger
}

// Option configures a Runner
type Option func(*Runner)

// New creates a new Runner with optional configuration
func New(opts ...Option) *Runner {
	r := &Runner{}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// WithLogger configures logging for command operations
func WithLogger(logger interfaces.Logger) Option {
	return func(r *Runner) {
		r.logger = logger
	}
}

// Combined runs a command given as a slice of strings and return a combined stdout and stderr string
func (r *Runner) Combined(cmdAndArgs ...string) (string, error) {
	return r.run(RunCombined, cmdAndArgs...)
}

// Separate runs a command given as a slice of strings and returns stdout and stderr separately in the same string
//
//goland:noinspection GoUnusedExportedFunction
func (r *Runner) Separate(cmdAndArgs ...string) (string, error) {
	return r.run(RunSeparate, cmdAndArgs...)
}

// Stdout runs a command given as a slice of strings and return stdout only
func (r *Runner) Stdout(cmdAndArgs ...string) (string, error) {
	return r.run(RunStdout, cmdAndArgs...)
}

// Stderr runs a command given as a slice of strings and return stderr only
//
//goland:noinspection GoUnusedExportedFunction
func (r *Runner) Stderr(cmdAndArgs ...string) (string, error) {
	return r.run(RunStderr, cmdAndArgs...)
}

// run a command given as a slice of strings and return output based on runType
func (r *Runner) run(runType int, cmdAndArgs ...string) (string, error) {
	var err error
	var out []byte
	var outStr string
	var stdout, stderr bytes.Buffer

	if len(cmdAndArgs) == 0 {
		return "", fmt.Errorf("no command provided")
	}

	// Log command execution (arguments redacted for security)
	if r.logger != nil {
		r.logger.Debugf(8350, "executing command: %s (arguments redacted)", cmdAndArgs[0])
	}

	// Set up the command
	cmd := exec.Command(cmdAndArgs[0], cmdAndArgs[1:]...)

	// Run it using the correct variant
	switch runType {
	case RunCombined:
		out, err = cmd.CombinedOutput()
		outStr = string(out)
	case RunStdout:
		out, err = cmd.Output()
		outStr = string(out)
	case RunStderr:
		cmd.Stdout = io.Discard
		cmd.Stderr = &stderr
		err = cmd.Run()
		outStr = stderr.String()
	case RunSeparate:
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err = cmd.Run()
		outStr = fmt.Sprintf("--- stdout ---\n%s\n\n--- stderr ---\n%s", stdout.String(), stderr.String())
	}

	// Capture the exit code
	exitCode := -1
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	// Log result
	if r.logger != nil {
		if err != nil {
			r.logger.Debugf(8351, "command failed: %s, exit code: %d, error: %v", cmdAndArgs[0], exitCode, err)
		} else {
			r.logger.Debugf(8352, "command succeeded: %s, exit code: %d", cmdAndArgs[0], exitCode)
		}
	}

	// Check for an error
	if err != nil {
		return outStr, fmt.Errorf("command %s failed with exit code %d: %s: %w",
			cmdAndArgs[0], exitCode, outStr, err)
	}

	// Check for edge case where err == nil and exitCode != 0
	if exitCode != 0 {
		return outStr, fmt.Errorf("command %s failed with exit code %d: %s",
			cmdAndArgs[0], exitCode, outStr)
	}

	return outStr, nil
}
