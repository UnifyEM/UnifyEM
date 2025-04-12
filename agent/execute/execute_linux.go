//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

//go:build linux

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

	// Ensure the file is executable
	if err = os.Chmod(file, 0755); err != nil {
		return fmt.Errorf("error setting executable bit on %s: %w", file, err)
	}

	cmd := exec.Command(file, args...)

	// Detach: new session, new process group
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setpgid: true,
	}

	// Redirect stdout and stderr to /dev/null for detachment, leave stdin unset
	devNull, err := os.OpenFile("/dev/null", os.O_RDWR, 0)
	if err == nil {
		cmd.Stdout = devNull
		cmd.Stderr = devNull
	}

	err = cmd.Start()
	if err != nil {
		_ = os.Remove(file)
		return fmt.Errorf("error executing %s: %w", file, err)
	}

	logger.Infof(8199, "Started %s with argument(s) %v", file, args)
	return nil
}
