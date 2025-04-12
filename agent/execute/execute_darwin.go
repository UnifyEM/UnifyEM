//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

//go:build darwin

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

	// Detach: new session, new process group (matches previous logic)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setpgid: true,
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		_ = os.Remove(file)
		return fmt.Errorf("error executing %s: %w", file, err)
	}

	logger.Infof(8199, "Started %s with argument(s) %v", file, args)
	return nil
}
