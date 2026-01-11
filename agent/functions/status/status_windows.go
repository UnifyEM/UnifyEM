/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package status

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows/registry"
)

var screenLockDelayValue string

func (h *Handler) osName() string {
	return "Windows"
}

func (h *Handler) osVersion() string {

	// Try to use PowerShell to get OS information
	out, err := exec.Command("powershell", "-Command",
		"Get-CimInstance Win32_OperatingSystem | Select-Object Caption, Version, BuildNumber | ConvertTo-Json").Output()
	if err == nil {
		output := strings.TrimSpace(string(out))
		// Parse the JSON output and extract the version information
		if strings.Contains(output, "Caption") && strings.Contains(output, "Version") {
			// Extract caption and version
			caption := ""
			version := ""

			if strings.Contains(output, "\"Caption\":") {
				captionParts := strings.Split(strings.Split(output, "\"Caption\":")[1], "\"")
				if len(captionParts) > 1 {
					caption = captionParts[1]
				}
			}

			if strings.Contains(output, "\"Version\":") {
				versionParts := strings.Split(strings.Split(output, "\"Version\":")[1], "\"")
				if len(versionParts) > 1 {
					version = versionParts[1]
				}
			}

			if caption != "" && version != "" {
				return caption + " (Version " + version + ")"
			} else if caption != "" {
				return caption
			}
		}
	}

	// If that fails, try the registry
	k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err == nil {
		defer func(k registry.Key) {
			_ = k.Close()
		}(k)

		productName, _, _ := k.GetStringValue("ProductName")
		displayVersion, _, _ := k.GetStringValue("DisplayVersion")

		if productName != "" {
			if displayVersion != "" {
				return productName + " " + displayVersion
			}
			return productName
		}
	}

	// If that files, try using the ver command as last resort
	out, err = exec.Command("cmd", "/c", "ver").Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}

	return "unknown"
}

func (h *Handler) firewall() string {
	out, err := exec.Command("netsh", "advfirewall", "show", "allprofiles").Output()
	if err != nil {
		return "unknown"
	}
	output := strings.ToLower(string(out))
	lines := strings.Split(output, "\n")
	stateOnCount := 0
	for _, line := range lines {
		if strings.Contains(line, "state") {
			state := strings.Fields(line)
			if len(state) > 1 && state[1] == "on" {
				stateOnCount++
			}
		}
	}

	if stateOnCount > 0 {
		return "yes"
	}
	return "no"
}

func (h *Handler) antivirus() string {

	for _, keyPath := range windowsAntivirusKeys {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.QUERY_VALUE)
		if err == nil {
			_ = k.Close()
			return "yes"
		}
	}

	return "no"
}

func (h *Handler) autoUpdates() string {
	noAutoUpdate, err := h.registryGetInt(registry.LOCAL_MACHINE, "SOFTWARE\\Policies\\Microsoft\\Windows\\WindowsUpdate\\AU", "NoAutoUpdate")
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return "yes"
		}
		return "unknown"
	}
	if noAutoUpdate == 1 {
		return "no"
	}
	return "yes"
}

func (h *Handler) fde() string {
	out, err := exec.Command("powershell", "Get-BitLockerVolume", "|", "Select-Object", "-ExpandProperty", "VolumeStatus").Output()
	if err != nil {
		return "unknown"
	}
	output := strings.TrimSpace(string(out))
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if !strings.Contains(line, "FullyEncrypted") && !strings.Contains(line, "EncryptionInProgress") {
			return "no"
		}
	}
	return "yes"
}

func (h *Handler) password() string {

	// Check AutoAdminLogon
	autoAdminLogon, err := h.registryGetString(registry.LOCAL_MACHINE, "SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion\\Winlogon", "AutoAdminLogon")
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return "unknown"
		}
	}
	if autoAdminLogon == "1" {
		return "no"
	}

	// Check non-admin auto-login
	defaultPassword, err := h.registryGetString(registry.LOCAL_MACHINE, "SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion\\Winlogon", "DefaultPassword")
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return "yes"
		}
		return "unknown"
	}
	if defaultPassword != "" {
		return "no"
	}

	return "yes"
}

func (h *Handler) screenLock() (string, error) {
	screenLockDelayValue = "0"

	screenSaverSecure, screenSaverTimeout, err := h.screenSaver()
	if err != nil {
		return "unknown", fmt.Errorf("error checking screen saver setting: %w", err)
	}

	info, err := GetScreenLockInfo()
	if err != nil {
		return "unknown", fmt.Errorf("error checking screen lock setting: %w", err)
	}

	// If a password is not required, there is no point considering the screen lock delay
	if !info.ConsoleLockDC {
		info.TimeoutDC = 0
	}

	if !info.ConsoleLockAC {
		info.TimeoutAC = 0
	}

	// Find the longest of the screen lock times (AC vs DC)
	var screenLockTimeout uint32 = 0
	if info.TimeoutAC > info.TimeoutDC {
		screenLockTimeout = info.TimeoutAC
	} else {
		screenLockTimeout = info.TimeoutDC
	}

	// There is no point comparing unless both are effective
	if screenSaverSecure && screenSaverTimeout > 0 && screenLockTimeout != 0 {

		// Both are enabled, so the shortest time gets the lock
		// Save the value for screenLockDelay() to use
		if screenSaverTimeout < screenLockTimeout {
			screenLockDelayValue = fmt.Sprintf("%d", screenSaverTimeout)
		} else {
			screenLockDelayValue = fmt.Sprintf("%d", screenLockTimeout)
		}
	}

	// Do we have an effective screen saver?
	if screenSaverSecure && screenSaverTimeout > 0 {
		return "yes", nil
	}

	// Do we have an effective screen lock on both AC and DC?
	if info.ConsoleLockAC && info.ConsoleLockDC && info.TimeoutAC > 0 && info.TimeoutDC > 0 {
		return "yes", nil
	}

	return "no", nil
}

func (h *Handler) screenLockDelay() string {
	// On Windows it is easiest to get this at the same time as screenLock() so it saves it
	return screenLockDelayValue
}

// screenSaver checks if the screen saver is enabled and if a password is required
func (h *Handler) screenSaver() (bool, uint32, error) {
	var secure = false
	var timeout uint32 = 0

	// Get the user's registry path - this is required because registry.CURRENT_USER would return the
	// service's context, not the user
	userPath, err := h.getUserRegistryKey()
	if err != nil {
		return false, 0, err
	}

	// Check if screen saver is configured to require a password
	secureValue, err := h.registryGetStringToInt(userPath, "Control Panel\\Desktop", "ScreenSaverIsSecure")
	if err != nil {
		return false, 0, fmt.Errorf("error checking for screen saver password: %w", err)
	}
	if secureValue == 1 {
		secure = true
	}

	// Check the screen saver timeout
	timeout, err = h.registryGetStringToInt(userPath, "Control Panel\\Desktop", "ScreenSaveTimeOut")
	if err != nil {
		return false, 0, fmt.Errorf("error checking screen saver timeout: %w", err)
	}

	// Check if the screen saver is set to (none)
	screenSaverValue, err := h.registryGetString(userPath, "Control Panel\\Desktop", "SCRNSAVE.EXE")
	if err != nil {
		return false, 0, fmt.Errorf("error checking screen saver executable: %w", err)
	}
	if screenSaverValue == "" || screenSaverValue == "(none)" {
		return false, 0, nil
	}

	if timeout > 60 {
		timeout = timeout / 60
	} else {
		timeout = 0
	}
	return secure, timeout, nil
}

func (h *Handler) lastUser() string {

	// Check the currently logged-in user
	out, err := exec.Command("query", "user").Output()
	if err == nil {
		output := strings.TrimSpace(string(out))
		lines := strings.Split(output, "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) > 0 {
				return fields[0]
			}
		}
	}

	// Check the last logged-in user from the registry
	val, err := h.registryGetString(registry.LOCAL_MACHINE, "SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Authentication\\LogonUI", "LastLoggedOnUser")
	if err != nil {
		return "unknown"
	}
	return val
}

func (h *Handler) bootTime() string {
	// Get system uptime
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getTickCount64 := kernel32.NewProc("GetTickCount64")
	ret, _, _ := getTickCount64.Call()
	uptime := time.Duration(ret) * time.Millisecond

	// Calculate boot time from uptime
	bTime := time.Now().Add(-uptime)
	return bTime.Format("2006-01-02T15:04:05-07:00")
}

// registryGetInt retrieves an integer value from the Windows registry
func (h *Handler) registryGetInt(key registry.Key, path string, name string) (int, error) {
	k, err := registry.OpenKey(key, path, registry.QUERY_VALUE)
	if err != nil {
		return 0, fmt.Errorf("error opening registry key: %w", err)
	}
	defer func(k registry.Key) {
		_ = k.Close()
	}(k)

	val, _, err := k.GetIntegerValue(name)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			// Registry key or value doesn't exist, assume it's not set
			return 0, nil
		}
		return 0, fmt.Errorf("error getting registry value: %w", err)
	}

	return int(val), nil
}

// registryGetString retrieves a string value from the Windows registry
func (h *Handler) registryGetString(key registry.Key, path string, name string) (string, error) {
	k, err := registry.OpenKey(key, path, registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("error opening registry key: %w", err)
	}
	defer func(k registry.Key) {
		_ = k.Close()
	}(k)

	val, _, err := k.GetStringValue(name)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			// Registry key or value doesn't exist, assume it's not set
			return "", nil
		}
		return "", fmt.Errorf("error getting registry value: %w", err)
	}

	return val, nil
}

// registryGetStringToInt retrieves a string value from the Windows registry and converts it to an integer
func (h *Handler) registryGetStringToInt(key registry.Key, path string, name string) (uint32, error) {
	strVal, err := h.registryGetString(key, path, name)
	if err != nil {
		return 0, err
	}

	// The string value may be empty or not set
	if strVal == "" {
		return 0, nil
	}

	uintVal, err := strconv.ParseUint(strVal, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("error converting string to uint32: %w", err)
	}

	return uint32(uintVal), nil
}

// getUserRegistryKey gets the registry path for the currently or last logged-in user
// This is used in place of registry.CURRENT_USER which would return the service's context
func (h *Handler) getUserRegistryKey() (registry.Key, error) {

	// Open the HKEY_USERS registry key
	usersKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Authentication\LogonUI`, registry.QUERY_VALUE)
	if err != nil {
		return 0, fmt.Errorf("error opening HKEY_USERS key: %w", err)
	}
	defer func(usersKey registry.Key) {
		_ = usersKey.Close()
	}(usersKey)

	// Get the last logged-on user SID
	lastLoggedOnUserSID, _, err := usersKey.GetStringValue("LastLoggedOnUserSID")
	if err != nil {
		return 0, fmt.Errorf("error getting LastLoggedOnUserSID: %w", err)
	}

	// Check if the user registry key exists
	userKeyPath := fmt.Sprintf(`%s`, lastLoggedOnUserSID)
	_, err = registry.OpenKey(registry.USERS, userKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return 0, fmt.Errorf("error checking user registry key: %w", err)
	}

	// Open the registry key for the user
	userKey, err := registry.OpenKey(registry.USERS, userKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return 0, fmt.Errorf("error opening user registry key: %w", err)
	}
	return userKey, nil
}

// checkServiceAccount is not implemented for Windows
func (h *Handler) checkServiceAccount() string {
	return "n/a"
}

// info returns platform-specific informational items
func (h *Handler) info() []string {
	return []string{}
}
