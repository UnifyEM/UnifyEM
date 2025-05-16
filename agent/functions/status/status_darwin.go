//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package status

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"howett.net/plist"
)

func (h *Handler) osName() string {
	return "macOS"
}

func (h *Handler) osVersion() string {
	out, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func (h *Handler) firewall() string {

	// Try plist first
	state, err := h.getPlistValue("/Library/Preferences/com.apple.alf", "globalstate")
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

func (h *Handler) antivirus() string {
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

func (h *Handler) autoUpdates() string {
	out, err := h.getPlistValue("/Library/Preferences/com.apple.SoftwareUpdate",
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

func (h *Handler) fde() string {
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

func (h *Handler) password() string {
	out, err := h.getPlistValue("/Library/Preferences/com.apple.loginwindow", "autoLoginUser")
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

func (h *Handler) screenLockDelay() string {
	username := h.lastUser()
	if username == "unknown" {
		return "unknown"
	}
	enabled, _, delay, err := h.getUserScreenSaverStatus(username)
	if err != nil {
		// fallback to AppleScript for current user context
		out, err2 := h.getAppleScript("tell application \"System Events\" to get delay interval of screen saver preferences")
		if err2 != nil {
			return "unknown"
		}
		return strings.TrimSpace(out)
	}
	if !enabled {
		return "0"
	}
	return fmt.Sprintf("%d", delay)
}

func (h *Handler) screenLock() (string, error) {
	username := h.lastUser()
	if username == "unknown" {
		return "unknown", fmt.Errorf("could not determine last user")
	}
	enabled, requirePassword, _, err := h.getUserScreenSaverStatus(username)
	if err != nil {
		// fallback to AppleScript for current user context
		out, err2 := h.getAppleScript("tell application \"System Events\" to get require password to wake of security preferences")
		if err2 != nil {
			h.logger.Errorf(2710, "error getting screen lock status from AppleScript: %s [%s]",
				err2.Error(), out)
			return "unknown", fmt.Errorf("error getting screen lock status from AppleScript: %w", err2)
		}
		if out == "true" {
			return "yes", nil
		}
		return "no", nil
	}
	if enabled && requirePassword {
		return "yes", nil
	}
	return "no", nil
}

// getUserScreenSaverStatus checks the screensaver/lock status for a given user
func (h *Handler) getUserScreenSaverStatus(username string) (enabled bool, requirePassword bool, delay int, err error) {
	usr, err := user.Lookup(username)
	if err != nil {
		return false, false, 0, fmt.Errorf("could not lookup user: %w", err)
	}

	// 1. Read idleTime from ByHost plist if present
	byHostPattern := filepath.Join(usr.HomeDir, "Library/Preferences/ByHost/com.apple.screensaver*.plist")
	byHostFiles, _ := filepath.Glob(byHostPattern)
	idleTime := 0
	askForPassword := 0
	askForPasswordDelay := 0
	if len(byHostFiles) > 0 {
		data, err := os.ReadFile(byHostFiles[0])
		if err == nil {
			var byHostData map[string]interface{}
			_, err = plist.Unmarshal(data, &byHostData)
			if err == nil {
				if v, ok := byHostData["idleTime"]; ok {
					switch t := v.(type) {
					case uint64:
						idleTime = int(t)
					case int64:
						idleTime = int(t)
					case int:
						idleTime = t
					case float64:
						idleTime = int(t)
					}
				}
				if v, ok := byHostData["askForPassword"]; ok {
					switch t := v.(type) {
					case uint64:
						askForPassword = int(t)
					case int64:
						askForPassword = int(t)
					case int:
						askForPassword = t
					case float64:
						askForPassword = int(t)
					}
				}
				if v, ok := byHostData["askForPasswordDelay"]; ok {
					switch t := v.(type) {
					case uint64:
						askForPasswordDelay = int(t)
					case int64:
						askForPasswordDelay = int(t)
					case int:
						askForPasswordDelay = t
					case float64:
						askForPasswordDelay = int(t)
					}
				}
			}
		}
	}

	// 2. Fallback: Read askForPassword and askForPasswordDelay from main plist if not found in ByHost
	if askForPassword == 0 || askForPasswordDelay == 0 {
		plistPath := filepath.Join(usr.HomeDir, "Library/Preferences/com.apple.screensaver.plist")
		data, err := os.ReadFile(plistPath)
		if err == nil {
			var plistData map[string]interface{}
			_, err = plist.Unmarshal(data, &plistData)
			if err == nil {
				if askForPassword == 0 {
					if v, ok := plistData["askForPassword"]; ok {
						switch t := v.(type) {
						case uint64:
							askForPassword = int(t)
						case int64:
							askForPassword = int(t)
						case int:
							askForPassword = t
						case float64:
							askForPassword = int(t)
						}
					}
				}
				if askForPasswordDelay == 0 {
					if v, ok := plistData["askForPasswordDelay"]; ok {
						switch t := v.(type) {
						case uint64:
							askForPasswordDelay = int(t)
						case int64:
							askForPasswordDelay = int(t)
						case int:
							askForPasswordDelay = t
						case float64:
							askForPasswordDelay = int(t)
						}
					}
				}
			}
		}
	}

	enabled = idleTime > 0
	delay = idleTime

	// If askForPassword is not set, try AppleScript as a last resort
	if askForPassword == 0 {
		script := `tell application "System Events" to get require password to wake of security preferences`
		out, err := h.runUserAppleScript(username, script)
		if err == nil {
			requirePassword = (out == "true")
			return enabled, requirePassword, delay, nil
		}
		h.logger.Warningf(2711, "Could not determine askForPassword for user %s, reporting as unknown", username)
		return enabled, false, delay, fmt.Errorf("askForPassword not set in any plist or via AppleScript")
	}

	requirePassword = askForPassword == 1
	return enabled, requirePassword, delay, nil
}

func (h *Handler) bootTime() string {
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

func (h *Handler) lastUser() string {
	out, err := exec.Command("defaults", "read", "/Library/Preferences/com.apple.loginwindow", "lastUserName").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

// getPlistValue retrieves the value associated with name from a plist at location
func (h *Handler) getPlistValue(location string, name string) (string, error) {
	value, err := exec.Command("defaults", "-currentHost", "read", location, name).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(value)), nil
}

/*
runUserAppleScript runs /usr/bin/osascript as the specified user (using sudo -u).
If username is empty, it falls back to the last user.
Returns the trimmed output or an error.
*/
func (h *Handler) runUserAppleScript(username, script string) (string, error) {
	if username == "" {
		username = h.lastUser()
	}
	if username == "unknown" {
		return "", fmt.Errorf("no user available to run AppleScript")
	}
	h.logger.Debugf(2712, "executing /bin/launchctl asuser %s sudo -u %s /usr/bin/osascript -e %s",
		username, username, "'"+script+"'")
	cmd := exec.Command("/bin/launchctl", "asuser", username, "sudo", "-u", username, "/usr/bin/osascript", "-e", "'"+script+"'")
	out, err := cmd.CombinedOutput()
	if err != nil {
		h.logger.Errorf(2709, "runUserAppleScript failed: %s [%s]",
			err.Error(), string(out))
		return "", fmt.Errorf("runUserAppleScript failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// getCurrentOrLastUser returns the currently logged-in user, or falls back to lastUser().
func (h *Handler) getCurrentOrLastUser() string {
	// Try "who" to get the console user
	out, err := exec.Command("/usr/bin/who").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) > 1 && fields[1] == "console" {
				return fields[0]
			}
		}
	}
	// Fallback to lastUser()
	return h.lastUser()
}

// getAppleScript runs an AppleScript as the current or last user (for backward compatibility)
func (h *Handler) getAppleScript(script string) (string, error) {
	username := h.getCurrentOrLastUser()
	return h.runUserAppleScript(username, script)
}
