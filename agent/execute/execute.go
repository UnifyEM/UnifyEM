//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package execute

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

func Execute(logger interfaces.Logger, file string, args []string) error {
	var err error

	// Execute the file with the supplied arguments
	cmd := exec.Command(file, args...)

	// Set platform-specific process attributes so that the child process can continue after
	// the parent process (which one) exists
	setProcessAttributes(cmd)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the command and do not wait for it to complete
	err = cmd.Start()
	if err != nil {
		_ = os.Remove(file)
		return fmt.Errorf("error executing %s: %w", file, err)
	}

	// Do not wait for the command to complete
	logger.Infof(8199, "Started %s with argument(s) %v", file, args)
	return nil
}
