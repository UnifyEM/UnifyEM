//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package status

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func osName() string {
	return "macOS"
}

func osVersion() string {
	out, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func firewall() string {

	// Try plist first
	state, err := getPlistValue("/Library/Preferences/com.apple.alf", "globalstate")
	if err == nil {
		if state == "1" {
			return "yes"
		}
		return "no"
	}

	// Try socketfilterfw
	out, err := exec.Command("/usr/libexec/ApplicationFirewall/socketfilterfw", "--getglobalstate").Output()
	if err != nil {
		return "unknown"
	}

	if strings.Contains(string(out), "Firewall is enabled") {
		return "yes"
	} else if strings.Contains(string(out), "Firewall is disabled") {
		return "no"
	} else {
		return "unknown"
	}
}

func antivirus() string {
	for _, path := range macAntivirusPaths {
		if _, err := exec.Command("test", "-e", path).Output(); err == nil {
			return "yes"
		}
	}

	out, err := exec.Command("ps", "aux").Output()
	if err != nil {
		return "unknown"
	}

	for _, process := range antivirusProcesses {
		if strings.Contains(string(out), process) {
			return "yes"
		}
	}

	return "no"
}

func autoUpdates() string {
	out, err := getPlistValue("/Library/Preferences/com.apple.SoftwareUpdate",
		"AutomaticallyInstallMacOSUpdates")
	if err != nil {
		return "unknown"
	}
	autoDownload := strings.TrimSpace(out)

	if autoDownload == "1" {
		return "yes"
	}
	return "no"
}

func fde() string {
	out, err := exec.Command("fdesetup", "status").Output()
	if err != nil {
		return "unknown"
	}
	status := strings.TrimSpace(string(out))
	if strings.Contains(status, "FileVault is On.") {
		return "yes"
	}
	return "no"
}

func password() string {
	out, err := getPlistValue("/Library/Preferences/com.apple.loginwindow", "autoLoginUser")
	if err != nil {
		// Check if the exit code indicates the key does not exist
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			if exitError.ExitCode() == 1 {
				return "yes"
			}
		}
		return "unknown"
	}
	autoLoginUser := strings.TrimSpace(out)
	if autoLoginUser == "" {
		return "yes"
	}
	return "no"
}

func screenLockDelay() string {
	out, err := getAppleScript("tell application \"System Events\" to get delay interval of screen saver preferences")
	if err != nil {
		return "unknown"
	}

	return strings.TrimSpace(out)
}

func screenLock() (string, error) {
	out, err := getAppleScript("tell application \"System Events\" to get require password to wake of security preferences")
	if err != nil {
		return "unknown", fmt.Errorf("error getting screen lock status: %w", err)
	}

	if out == "true" {
		return "yes", nil
	}
	return "no", nil
}

func bootTime() string {
	out, err := exec.Command("sysctl", "-n", "kern.boottime").Output()
	if err != nil {
		return "unknown"
	}
	bootTimeStr := strings.TrimSpace(string(out))
	bootTimeStr = strings.Trim(bootTimeStr, "{")
	bootTimeParts := strings.Split(bootTimeStr, ",")
	if len(bootTimeParts) < 2 {
		return "unknown"
	}

	// Extract the epoch from the string
	epoch := strings.TrimSpace(strings.Split(bootTimeParts[0], "=")[1])

	// Convert the epoch string to an int64
	epochInt, err := strconv.ParseInt(epoch, 10, 64)
	if err != nil {
		return "unknown"
	}

	// Convert the epoch int64 to a time.Time
	bootTimeInt := time.Unix(epochInt, 0)

	// Format the time.Time to a string
	return bootTimeInt.Format("2006-01-02T15:04:05-07:00")
}

func lastUser() string {
	out, err := exec.Command("defaults", "read", "/Library/Preferences/com.apple.loginwindow", "lastUserName").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

// getPlistValue retrieves the value associated with name from a plist at location
func getPlistValue(location string, name string) (string, error) {
	value, err := exec.Command("defaults", "-currentHost", "read", location, name).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(value)), nil
}

// getAppleScript runs an AppleScript and returns the output
func getAppleScript(script string) (string, error) {
	out, err := exec.Command("/usr/bin/osascript", "-e", script).Output()
	if err != nil {
		// Wait 5 seconds and try again
		time.Sleep(5 * time.Second)
		out, err = exec.Command("/usr/bin/osascript", "-e", script).Output()
		if err != nil {
			return "", err
		}
	}
	return strings.TrimSpace(string(out)), nil
}
