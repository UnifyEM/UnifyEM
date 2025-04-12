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

	// systemd has a nasty habbit of killing child processes when the parent dies
	// so we need to use systemd-run to start the process in a new scope
	newArgs := []string{"--scope", "--quiet"}
	newArgs = append(newArgs, file)
	newArgs = append(newArgs, args...)
	cmd := exec.Command("systemd-run", newArgs...)

	// Detach from parent
	//cmd.SysProcAttr = &syscall.SysProcAttr{
	//	Setpgid: true,
	//}

	// Redirect stdin, stdout, stderr to /dev/null
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("Failed to open /dev/null: %w", err)
	}
	defer devNull.Close()

	cmd.Stdin = devNull
	cmd.Stdout = devNull
	cmd.Stderr = devNull

	err = cmd.Start()
	if err != nil {
		_ = os.Remove(file)
		return fmt.Errorf("error executing %s: %w", file, err)
	}

	logger.Infof(8199, "Started %s with argument(s) %v", file, args)
	return nil
}
