//go:build windows

/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

// Code for Windows

package osActions

import (
	"fmt"
	"os"
	"strings"

	"github.com/StackExchange/wmi"

	"github.com/UnifyEM/UnifyEM/common/runCmd"
	"github.com/UnifyEM/UnifyEM/common/schema"
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
func (a *Actions) lockUser(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	uq, err := safeUsername(username)
	if err != nil {
		return err
	}

	_, err = runCmd.Combined("net", "user", uq, "/active:no")
	if err != nil {
		return fmt.Errorf("failed to lock user %s: %w", uq, err)
	}
	return nil
}

// unlockUser enables the user account to allow access
func (a *Actions) unlockUser(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	uq, err := safeUsername(username)
	if err != nil {
		return err
	}

	_, err = runCmd.Combined("net", "user", uq, "/ACTIVE:yes")
	if err != nil {
		return fmt.Errorf("failed to unlock user %s: %w", uq, err)
	}
	return nil
}

// setPassword sets the password for the specified user
func (a *Actions) setPassword(username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("username and password cannot be empty")
	}

	uq, err := safeUsername(username)
	if err != nil {
		return err
	}

	pq, err := safePassword(password)
	if err != nil {
		return err
	}

	_, err = runCmd.Combined("net", "user", uq, pq)
	if err != nil {
		return fmt.Errorf("failed to set password for user %s: %w", uq, err)
	}
	return nil
}

// addUser creates a new user and sets their password
func (a *Actions) addUser(username, password string, admin bool) error {
	if username == "" || password == "" {
		return fmt.Errorf("username and password cannot be empty")
	}

	uq, err := safeUsername(username)
	if err != nil {
		return err
	}

	pq, err := safePassword(password)
	if err != nil {
		return err
	}

	// Create the user and set the password
	_, err = runCmd.Combined("net", "user", uq, pq, "/ADD")
	if err != nil {
		return fmt.Errorf("failed to create user %s: %w", uq, err)
	}

	// Add the user's password to allow boot drive bitlocker access
	strippedPQ := strings.ReplaceAll(pq, "'", "''") // escape single quotes
	_, err = runCmd.Combined(
		"powershell",
		"-Command",
		fmt.Sprintf(
			"Add-BitLockerKeyProtector -MountPoint 'C:' -PasswordProtector -Password (ConvertTo-SecureString '%s' -AsPlainText -Force)",
			strippedPQ,
		),
	)
	if err != nil {
		if strings.Contains(err.Error(), "BitLocker is not enabled") {
			// Handle the case where BitLocker is not enabled
			a.logger.Warningf(8401, "BitLocker is not enabled, not adding user %s to BitLocker", uq)
		} else {
			return fmt.Errorf("failed adding password to BitLocker: %w", err)
		}
	}

	// Check if the user should be an admin
	if admin {
		return a.setAdmin(username, true)
	}
	return nil
}

func (a *Actions) setAdmin(username string, admin bool) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	uq, err := safeUsername(username)
	if err != nil {
		return err
	}

	if admin {
		err = a.addToGroup(uq, "Administrators")
		if err != nil {
			return fmt.Errorf("failed to add user %s from Administators group: %w", uq, err)
		}
	} else {
		err = a.removeFromGroup(uq, "Administrators")
		if err != nil {
			return fmt.Errorf("failed to remove user %s from Administators group: %w", uq, err)
		}
		// Just a best practice, but not really needed
		_ = a.addToGroup(uq, "User")
	}
	return nil
}

func (a *Actions) addToGroup(user, group string) error {
	out, err := runCmd.Combined("net", "localgroup", group, user, "/ADD")
	if err != nil {
		if strings.Contains(string(out), "is already a member") {
			// Consider this a success
			return nil
		}

		return fmt.Errorf("failed to set user %s as %s: %w", user, group, err)
	}
	return nil
}

func (a *Actions) removeFromGroup(user, group string) error {
	out, err := runCmd.Combined("net", "localgroup", group, user, "/DELETE")
	if err != nil {
		if strings.Contains(string(out), "is not a member") {
			// Consider this a success
			return nil
		}

		return fmt.Errorf("failed to set user %s as %s: %w", user, group, err)
	}
	return nil
}

// deleteUser removes a user from the system
func (a *Actions) deleteUser(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	uq, err := safeUsername(username)
	if err != nil {
		return err
	}

	// Delete the user
	_, err = runCmd.Combined("net", "user", uq, "/DELETE")
	if err != nil {
		return fmt.Errorf("failed to delete user %s: %w", uq, err)
	}

	return nil
}
