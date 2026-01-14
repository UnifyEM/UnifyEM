//go:build windows

/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package osActions

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"github.com/StackExchange/wmi"

	"github.com/UnifyEM/UnifyEM/common/crypto"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

// Windows constants for LogonUser
const (
	LOGON32_LOGON_NETWORK    = 3
	LOGON32_PROVIDER_DEFAULT = 0
)

func (a *Actions) getUsers() (schema.DeviceUserList, error) {
	var users schema.DeviceUserList

	// Retrieve the local hostname
	hostname, err := os.Hostname()
	if err != nil {
		return schema.DeviceUserList{}, fmt.Errorf("failed to get hostname: %w", err)
	}

	a.logger.Debugf(8404, "hostname: %s", hostname)

	// Get all users using WMI
	var wmiUsers []struct {
		Name         string
		Domain       string
		Disabled     bool
		LocalAccount bool
	}

	query := "SELECT Name, Domain, Disabled, LocalAccount FROM Win32_UserAccount"
	err = wmi.Query(query, &wmiUsers)
	if err != nil {
		return schema.DeviceUserList{}, fmt.Errorf("WMI query failed: %w", err)
	}

	// Iterate through the list of users and add the local users to the list
	for _, wmiUser := range wmiUsers {
		a.logger.Debugf(8405, "Found user: %s Domain: %s Disabled: %t Local: %t", wmiUser.Name, wmiUser.Domain, wmiUser.Disabled, wmiUser.LocalAccount)
		if !wmiUser.LocalAccount {
			a.logger.Debugf(8406, "Skipping non-local account user: %s Domain: %s", wmiUser.Name, wmiUser.Domain)
			continue
		}
		users.Users = append(users.Users, schema.DeviceUser{Domain: wmiUser.Domain, Name: wmiUser.Name, Disabled: wmiUser.Disabled, Administrator: false})
	}

	// Get the list of administrators using WMI
	var adminList []struct {
		PartComponent string
	}

	adminQuery := fmt.Sprintf(`SELECT PartComponent FROM Win32_GroupUser WHERE GroupComponent="Win32_Group.Name='Administrators',Domain='%s'"`, hostname)
	err = wmi.Query(adminQuery, &adminList)
	if err != nil {
		return schema.DeviceUserList{}, fmt.Errorf("WMI query failed: %w", err)
	}

	// Log and parse
	a.logger.Debugf(8407, "Administrators: %v", adminList)
	admins, err := parseAdminList(adminList)

	// Iterate through the list of users and mark the administrators
	for i, user := range users.Users {
		key := fmt.Sprintf("%s\\%s", user.Domain, user.Name)
		if _, exists := admins[key]; exists {
			users.Users[i].Administrator = true
		}
	}

	// Check for any administrator that is not in the list of users
	// This is common with domain users
	for admin := range admins {
		found := false
		for _, user := range users.Users {
			if fmt.Sprintf("%s\\%s", user.Domain, user.Name) == admin {
				found = true
				break
			}
		}
		if !found {
			parts := strings.Split(admin, "\\")
			a.logger.Debugf(8408, "Adding admin user: %s Domain: %s Disabled: %t Local: %t", parts[1], parts[0], false, true)
			users.Users = append(users.Users, schema.DeviceUser{Domain: parts[0], Name: parts[1], Disabled: false, Administrator: true})
		}
	}
	return users, nil
}

func parseAdminList(adminList []struct{ PartComponent string }) (map[string]struct{}, error) {
	admins := make(map[string]struct{})
	for _, admin := range adminList {
		// Example PartComponent: \\C3PO\root\cimv2:Win32_UserAccount.Domain="C3PO",Name="Administrator"
		parts := strings.Split(admin.PartComponent, ":Win32_UserAccount.Domain=")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid PartComponent format: %s", admin.PartComponent)
		}
		domainAndName := parts[1]
		domainAndName = strings.Trim(domainAndName, "\"")
		domainAndNameParts := strings.Split(domainAndName, "\",Name=\"")
		if len(domainAndNameParts) < 2 {
			return nil, fmt.Errorf("invalid domain and name format: %s", domainAndName)
		}
		domain := domainAndNameParts[0]
		name := strings.Trim(domainAndNameParts[1], "\"")
		admins[fmt.Sprintf("%s\\%s", domain, name)] = struct{}{}
	}
	return admins, nil
}

// lockUser disables the user account to deny access
func (a *Actions) lockUser(userInfo UserInfo) error {
	if userInfo.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	_, err := a.runner.Combined("net", "user", userInfo.Username, "/ACTIVE:no")
	if err != nil {
		return fmt.Errorf("failed to lock user %s: %w", userInfo.Username, err)
	}
	return nil
}

// unlockUser enables the user account to allow access
func (a *Actions) unlockUser(userInfo UserInfo) error {
	if userInfo.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	_, err := a.runner.Combined("net", "user", userInfo.Username, "/ACTIVE:yes")
	if err != nil {
		return fmt.Errorf("failed to unlock user %s: %w", userInfo.Username, err)
	}
	return nil
}

// setPassword sets the password for the specified user
func (a *Actions) setPassword(userInfo UserInfo) error {

	if userInfo.Username == "" || userInfo.Password == "" {
		return fmt.Errorf("username and password are required")
	}

	_, err := a.runner.Combined("net", "user", userInfo.Username, userInfo.Password)
	if err != nil {
		return fmt.Errorf("failed to set password for user %s: %w", userInfo.Username, err)
	}
	return nil
}

// addUser creates a new user and sets their password
func (a *Actions) addUser(userInfo UserInfo) error {

	if userInfo.Username == "" || userInfo.Password == "" {
		return fmt.Errorf("username and password are required")
	}

	// Create the user and set the password
	_, err := a.runner.Combined("net", "user", userInfo.Username, userInfo.Password, "/ADD")
	if err != nil {
		return fmt.Errorf("failed to create user %s: %w", userInfo.Username, err)
	}

	// Add the user's password to allow boot drive bitlocker access
	// Escape PowerShell special characters: single quotes, backticks, dollar signs
	escapedPW := escapePowerShellString(userInfo.Password)
	_, err = a.runner.Combined(
		"powershell",
		"-Command",
		fmt.Sprintf(
			"Add-BitLockerKeyProtector -MountPoint 'C:' -PasswordProtector -Password (ConvertTo-SecureString '%s' -AsPlainText -Force)",
			escapedPW,
		),
	)
	if err != nil {
		if strings.Contains(err.Error(), "BitLocker is not enabled") {
			// Handle the case where BitLocker is not enabled
			a.logger.Warningf(8401, "BitLocker is not enabled, not adding user %s to BitLocker", userInfo.Username)
		} else {
			return fmt.Errorf("failed adding password to BitLocker: %w", err)
		}
	}

	// Check if the user should be an admin
	if userInfo.Admin {
		return a.setAdmin(userInfo)
	}
	return nil
}

func (a *Actions) setAdmin(userInfo UserInfo) error {
	if userInfo.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if userInfo.Admin {
		err := a.addToGroup(userInfo.Username, "Administrators")
		if err != nil {
			return fmt.Errorf("failed to add user %s from Administrators group: %w", userInfo.Username, err)
		}
	} else {
		err := a.removeFromGroup(userInfo.Username, "Administrators")
		if err != nil {
			return fmt.Errorf("failed to remove user %s from Administrators group: %w", userInfo.Username, err)
		}
		// Just a best practice, but not really needed
		_ = a.addToGroup(userInfo.Username, "User")
	}
	return nil
}

func (a *Actions) addToGroup(user, group string) error {
	out, err := a.runner.Combined("net", "localgroup", group, user, "/ADD")
	if err != nil {
		if strings.Contains(string(out), "is already a member") {
			// Consider this a success
			return nil
		}

		return fmt.Errorf("failed to add user %s to %s: %w", user, group, err)
	}
	return nil
}

func (a *Actions) removeFromGroup(user, group string) error {
	out, err := a.runner.Combined("net", "localgroup", group, user, "/DELETE")
	if err != nil {
		if strings.Contains(string(out), "is not a member") {
			// Consider this a success
			return nil
		}

		return fmt.Errorf("failed to remove user %s from %s: %w", user, group, err)
	}
	return nil
}

// deleteUser removes a user from the system
func (a *Actions) deleteUser(userInfo UserInfo) error {
	if userInfo.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Delete the user
	_, err := a.runner.Combined("net", "user", userInfo.Username, "/DELETE")
	if err != nil {
		return fmt.Errorf("failed to delete user %s: %w", userInfo.Username, err)
	}

	return nil
}

// userExists checks if a user exists on the system
func (a *Actions) userExists(username string) (bool, error) {
	if username == "" {
		return false, fmt.Errorf("username cannot be empty")
	}

	out, err := a.runner.Combined("net", "user", username)
	if err != nil {
		// Check if the error is because the user doesn't exist
		outStr := string(out)
		if strings.Contains(outStr, "not found") || strings.Contains(outStr, "could not be found") {
			return false, nil
		}
		// Some other error occurred
		return false, fmt.Errorf("failed to check if user %s exists: %w", username, err)
	}

	return true, nil
}

func (a *Actions) testCredentials(user string, pass string) error {
	if user == "" || pass == "" {
		return fmt.Errorf("username and password are required")
	}

	// Load advapi32.dll and get LogonUserW function
	advapi32 := syscall.NewLazyDLL("advapi32.dll")
	logonUser := advapi32.NewProc("LogonUserW")

	// Convert strings to UTF16 pointers
	userPtr, err := syscall.UTF16PtrFromString(user)
	if err != nil {
		return fmt.Errorf("failed to convert username: %w", err)
	}

	passPtr, err := syscall.UTF16PtrFromString(pass)
	if err != nil {
		return fmt.Errorf("failed to convert password: %w", err)
	}

	// Use "." for local domain
	domainPtr, err := syscall.UTF16PtrFromString(".")
	if err != nil {
		return fmt.Errorf("failed to convert domain: %w", err)
	}

	var token uintptr

	// Call LogonUserW
	ret, _, _ := logonUser.Call(
		uintptr(unsafe.Pointer(userPtr)),   // username
		uintptr(unsafe.Pointer(domainPtr)), // domain
		uintptr(unsafe.Pointer(passPtr)),   // password
		uintptr(LOGON32_LOGON_NETWORK),     // logon type
		uintptr(LOGON32_PROVIDER_DEFAULT),  // logon provider
		uintptr(unsafe.Pointer(&token)),    // token handle
	)

	// Close the token handle if logon succeeded
	if ret != 0 && token != 0 {
		syscall.CloseHandle(syscall.Handle(token))
		return nil
	}

	// Logon failed
	return fmt.Errorf("authentication failed for user %s: invalid credentials", user)
}

// refreshServiceAccount generates a new password for the service account and ensures it's an administrator
// Returns the new password on success
func (a *Actions) refreshServiceAccount(userInfo UserInfo) (string, error) {
	if userInfo.Username == "" {
		return "", fmt.Errorf("username is required")
	}

	// Ensure the user exists
	exists, err := a.userExists(userInfo.Username)
	if err != nil {
		return "", fmt.Errorf("failed to check if user exists: %w", err)
	}
	if !exists {
		return "", fmt.Errorf("user %s does not exist", userInfo.Username)
	}

	// Ensure the user is an administrator
	userInfo.Admin = true
	err = a.setAdmin(userInfo)
	if err != nil {
		return "", fmt.Errorf("failed to set admin status for user %s: %w", userInfo.Username, err)
	}

	// Generate a new random password
	newPassword := crypto.RandomPassword()

	// Set the new password using net user (runs as SYSTEM, no old password needed)
	_, err = a.runner.Combined("net", "user", userInfo.Username, newPassword)
	if err != nil {
		return "", fmt.Errorf("failed to change password for user %s: %w", userInfo.Username, err)
	}

	return newPassword, nil
}

// escapePowerShellString escapes special characters for use in PowerShell single-quoted strings
func escapePowerShellString(s string) string {
	// In PowerShell single-quoted strings, only single quotes need escaping (doubled)
	// Backticks, dollar signs, etc. are literal in single quotes
	return strings.ReplaceAll(s, "'", "''")
}
