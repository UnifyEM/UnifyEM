//go:build windows

/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package execute

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

func Execute(logger interfaces.Logger, file string, args []string) error {
	var err error

	// Execute the file with the supplied arguments
	cmd := exec.Command(file, args...)

	// Set process attributes so that the child process can continue after the parent process exists
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	// Redirect stdin, stdout, stderr to NUL
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to open NUL: %w", err)
	}
	defer func(devNull *os.File) {
		_ = devNull.Close()
	}(devNull)

	cmd.Stdin = devNull
	cmd.Stdout = devNull
	cmd.Stderr = devNull

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
