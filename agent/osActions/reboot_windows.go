//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

// Code for windows
//go:build windows

package osActions

import (
	"errors"
	"fmt"
	"golang.org/x/sys/windows"
	"os/exec"
)

//goland:noinspection GoSnakeCaseUsage
const SE_SHUTDOWN_NAME = "SeShutdownPrivilege"

// shutdownOrReboot attempts a clean shutdown/reboot using InitiateSystemShutdownExW,
// and falls back to ExitWindowsEx or shutdown command if necessary.
func (a *Actions) shutdownOrReboot(reboot bool) error {
	var err error

	err = attemptInitiateSystemShutdownEx(reboot)
	if err != nil {
		a.logger.Errorf(8304, "InitiateSystemShutdownEx failed: %s", err.Error())
	} else {
		return nil
	}

	err = attemptExitWindowsEx(reboot)
	if err != nil {
		a.logger.Errorf(8305, "ExitWindowsEx failed: %s", err.Error())
	} else {
		return nil
	}

	err = attemptShutdownCommand(reboot)
	if err != nil {
		a.logger.Errorf(8306, "Shutdown command failed: %s", err.Error())
	} else {
		return nil
	}
	//goland:noinspection GoErrorStringFormat
	return errors.New("Windows API calls and shutdown command failed")
}

// enableShutdownPrivilege enables the SE_SHUTDOWN_NAME privilege for the current process
func enableShutdownPrivilege() error {
	var token windows.Token

	processHandle := windows.CurrentProcess()
	err := windows.OpenProcessToken(processHandle, windows.TOKEN_ADJUST_PRIVILEGES|windows.TOKEN_QUERY, &token)
	if err != nil {
		return fmt.Errorf("OpenProcessToken: %w", err)
	}
	defer func(token windows.Token) {
		_ = token.Close()
	}(token)

	var luid windows.LUID
	err = windows.LookupPrivilegeValue(nil, windows.StringToUTF16Ptr(SE_SHUTDOWN_NAME), &luid)
	if err != nil {
		return fmt.Errorf("LookupPrivilegeValue: %w", err)
	}

	tp := windows.Tokenprivileges{
		PrivilegeCount: 1,
		Privileges: [1]windows.LUIDAndAttributes{
			{
				Luid:       luid,
				Attributes: windows.SE_PRIVILEGE_ENABLED,
			},
		},
	}

	err = windows.AdjustTokenPrivileges(token, false, &tp, 0, nil, nil)
	if err != nil {
		return fmt.Errorf("AdjustTokenPrivileges: %w", err)
	}

	return nil
}

func attemptInitiateSystemShutdownEx(reboot bool) error {
	if err := enableShutdownPrivilege(); err != nil {
		return fmt.Errorf("failed to enable shutdown privilege: %w", err)
	}

	advapi32 := windows.NewLazySystemDLL("advapi32.dll")
	proc := advapi32.NewProc("InitiateSystemShutdownExW")

	computerName := uintptr(0) // Local machine
	message := uintptr(0)      // No custom message
	timeout := uintptr(0)      // No delay
	forceAppsClosed := uintptr(1)
	rebootAfterShutdown := uintptr(0)
	if reboot {
		rebootAfterShutdown = uintptr(1)
	}
	reason := uintptr(0x80000000) // Major: Other, Minor: Unplanned

	r, _, err := proc.Call(computerName, message, timeout, forceAppsClosed, rebootAfterShutdown, reason)
	if r == 0 {
		return fmt.Errorf("InitiateSystemShutdownExW failed: %w", err)
	}
	return nil
}

// Attempt to use ExitWindowsEx
func attemptExitWindowsEx(reboot bool) error {
	if err := enableShutdownPrivilege(); err != nil {
		return fmt.Errorf("failed to enable shutdown privilege: %w", err)
	}

	user32 := windows.NewLazySystemDLL("user32.dll")
	proc := user32.NewProc("ExitWindowsEx")

	//goland:noinspection GoSnakeCaseUsage,GoUnusedConst
	const (
		EWX_LOGOFF   = 0x00000000
		EWX_SHUTDOWN = 0x00000001
		EWX_REBOOT   = 0x00000002
		EWX_FORCE    = 0x00000004
	)

	var flags uintptr
	if reboot {
		flags = EWX_REBOOT | EWX_FORCE
	} else {
		flags = EWX_SHUTDOWN | EWX_FORCE
	}

	r, _, err := proc.Call(flags, 0)
	if r == 0 {
		return fmt.Errorf("ExitWindowsEx failed: %w", err)
	}
	return nil
}

// Attempt to use the `shutdown` command
func attemptShutdownCommand(reboot bool) error {
	cmdArgs := []string{"/s", "/t", "0"} // Default: Shutdown
	if reboot {
		cmdArgs = []string{"/r", "/t", "0"} // Reboot
	}

	cmd := exec.Command("shutdown", cmdArgs...)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("shutdown command failed: %w", err)
	}
	return nil
}
