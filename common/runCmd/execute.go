/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package runCmd

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Combined runs a command given as a slice of strings and return a combined stdout and stderr string
func Combined(cmdAndArgs ...string) ([]byte, error) {
	if len(cmdAndArgs) == 0 {
		return []byte{}, fmt.Errorf("no command provided")
	}

	cmd := exec.Command(cmdAndArgs[0], cmdAndArgs[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {

		// If the process started but exited non-zero, wrap with exit code & output
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return []byte{}, fmt.Errorf(
				"command %q failed with exit code %d: %s: %w",
				strings.Join(cmdAndArgs, " "),
				exitErr.ExitCode(),
				string(out),
				err,
			)
		}

		// Other errors (e.g. executable not found)
		return []byte{}, fmt.Errorf(
			"failed to run %q: %w",
			strings.Join(cmdAndArgs, " "),
			err,
		)
	}

	return out, nil
}

// Stdout runs a command given as a slice of strings and return stdout only (stderr goes to os.Stderr)
func Stdout(cmdAndArgs ...string) ([]byte, error) {
	if len(cmdAndArgs) == 0 {
		return []byte{}, fmt.Errorf("no command provided")
	}

	cmd := exec.Command(cmdAndArgs[0], cmdAndArgs[1:]...)
	out, err := cmd.Output()
	if err != nil {

		// If the process started but exited non-zero, wrap with exit code & output
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return []byte{}, fmt.Errorf(
				"command %q failed with exit code %d: %s: %w",
				strings.Join(cmdAndArgs, " "),
				exitErr.ExitCode(),
				string(out),
				err,
			)
		}

		// Other errors (e.g. executable not found)
		return []byte{}, fmt.Errorf(
			"failed to run %q: %w",
			strings.Join(cmdAndArgs, " "),
			err,
		)
	}

	return out, nil
}
