//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

// Code for windows
//go:build windows

package osActions

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/StackExchange/wmi"

	"github.com/UnifyEM/UnifyEM/common/schema"
)

func (a *Actions) getUsers() (schema.DeviceUserList, error) {
	var users schema.DeviceUserList

	// Retrieve the local hostname
	hostname, err := os.Hostname()
	if err != nil {
		return schema.DeviceUserList{}, fmt.Errorf("failed to get hostname: %w", err)
	}

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
		if !wmiUser.LocalAccount {
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

	cmd := exec.Command("net", "user", uq, "/active:no")
	err = cmd.Run()
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

	cmd := exec.Command("net", "user", uq, "/active:yes")
	err = cmd.Run()
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

	cmd := exec.Command("net", "user", uq, pq)
	err = cmd.Run()
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
	cmd := exec.Command("net", "user", uq, pq, "/add")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to create user %s: %w", uq, err)
	}

	// Add the user to the "Users" group
	cmd = exec.Command("net", "localgroup", "Users", uq, "/add")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to add user %s to Users group: %w", uq, err)
	}

	// Get the domain and RID for the user
	cmd = exec.Command("wmic", "useraccount", "where", fmt.Sprintf("name='%s'", username), "get", "sid")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get SID for user %s: %w", uq, err)
	}

	sid := strings.TrimSpace(strings.Split(string(output), "\n")[1])
	domainUser := strings.Join(strings.Split(sid, "-")[:5], "-")
	rid := strings.Split(sid, "-")[5]

	// Add the user to the list of users who can unlock the drive with BitLocker
	cmd = exec.Command("manage-bde", "-protectors", "-add", "C:", "-sid", fmt.Sprintf("%s-%s", domainUser, rid))
	err = cmd.Run()
	if err != nil {
		if strings.Contains(err.Error(), "BitLocker is not enabled") {
			// Handle the case where BitLocker is not enabled
			fmt.Printf("BitLocker is not enabled, skipping adding user %s to BitLocker\n", uq)
		} else {
			return fmt.Errorf("failed to add user %s to BitLocker: %w", uq, err)
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

	var group string
	if admin {
		group = "Administrators"
	} else {
		group = "Users"
	}

	cmd := exec.Command("net", "localgroup", group, uq, "/add")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to set user %s as %s: %w", uq, group, err)
	}
	return nil
}
