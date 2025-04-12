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

	cmd := exec.Command(file, args...)

	// Detach: new session, new process group
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setpgid: true,
	}

	// Redirect stdio to /dev/null for full detachment
	devNull, err := os.OpenFile("/dev/null", os.O_RDWR, 0)
	if err == nil {
		cmd.Stdin = devNull
		cmd.Stdout = devNull
		cmd.Stderr = devNull
	} else {
		cmd.Stdin = nil
	}

	err = cmd.Start()
	if err != nil {
		_ = os.Remove(file)
		return fmt.Errorf("error executing %s: %w", file, err)
	}

	logger.Infof(8199, "Started %s with argument(s) %v", file, args)
	return nil
}
