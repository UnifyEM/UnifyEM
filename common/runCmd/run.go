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
)

const (
	RunCombined int = iota
	RunSeparate
	RunStdout
	RunStderr
)

// Combined runs a command given as a slice of strings and return a combined stdout and stderr string
func Combined(cmdAndArgs ...string) (string, error) {
	return run(RunCombined, cmdAndArgs...)
}

// Separate runs a command given as a slice of strings and returns stdout and stderr separately in the same tring
//
//goland:noinspection GoUnusedExportedFunction
func Separate(cmdAndArgs ...string) (string, error) {
	return run(RunSeparate, cmdAndArgs...)
}

// Stdout runs a command given as a slice of strings and return stdout only
func Stdout(cmdAndArgs ...string) (string, error) {
	return run(RunStdout, cmdAndArgs...)
}

// Stderr runs a command given as a slice of strings and return stdout only
//
//goland:noinspection GoUnusedExportedFunction
func Stderr(cmdAndArgs ...string) (string, error) {
	return run(RunStderr, cmdAndArgs...)
}

// run a command given as a slice of strings and return a combined stdout and stderr string
func run(runType int, cmdAndArgs ...string) (string, error) {
	var err error
	var out []byte
	var outStr string
	var stdout, stderr bytes.Buffer

	if len(cmdAndArgs) == 0 {
		return "", fmt.Errorf("no command provided")
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
