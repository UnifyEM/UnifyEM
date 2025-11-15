//go:build darwin

/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

// Code for macOS
package osActions

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/UnifyEM/UnifyEM/common/runCmd"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

func (a *Actions) getUsers() (schema.DeviceUserList, error) {

	// Get a list of users
	output, err := runCmd.Stdout("dscl", ".", "list", "/Users")
	if err != nil {
		return schema.DeviceUserList{}, err
	}

	// Get the admin group members
	admins, err := getAdminGroupMembers()
	if err != nil {
		return schema.DeviceUserList{}, err
	}

	// Parse the output
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	var users schema.DeviceUserList
	for scanner.Scan() {
		user := scanner.Text()

		if len(user) < 1 {
			continue
		}

		// Ignore the daemon users
		if strings.HasPrefix(user, "_") {
			continue
		}

		// Get admin status
		admin := false
		if _, isAdmin := admins[user]; isAdmin {
			admin = true
		}

		users.Users = append(users.Users,
			schema.DeviceUser{
				Name:          user,
				Domain:        "",
				Administrator: admin,
				Disabled:      !canUserLogin(user),
			},
		)
	}

	if err = scanner.Err(); err != nil {
		return schema.DeviceUserList{}, err
	}
	return users, nil
}

// getAdminGroupMembers retrieves the members of the admin group
func getAdminGroupMembers() (map[string]struct{}, error) {
	output, err := runCmd.Stdout("dscl", ".", "read", "/Groups/admin", "GroupMembership")
	if err != nil {
		return nil, err
	}

	members := strings.Fields(string(output))
	admins := make(map[string]struct{}, len(members))
	for _, member := range members {
		admins[member] = struct{}{}
	}
	return admins, nil
}

// canUserLogin checks if the user has a valid shell
func canUserLogin(username string) bool {
	if username == "" {
		return false
	}

	output, err := runCmd.Stdout("dscl", ".", "-read", "/Users/"+username, "UserShell")
	if err != nil {
		return false
	}

	shell := strings.TrimSpace(strings.Split(string(output), ":")[1])
	if shell == "/usr/sbin/uucico" {
		return false
	}
	return shell != "/usr/bin/false" && shell != "/usr/sbin/nologin"
}

// lockUser changes the user's shell to /usr/bin/false to deny access
func (a *Actions) lockUser(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	uq, err := safeUsername(username)
	if err != nil {
		return err
	}

	_, err = runCmd.Combined("dscl", ".", "-change", fmt.Sprintf("/Users/%s", uq), "UserShell", "/bin/bash", "/usr/bin/false")
	if err != nil {
		return fmt.Errorf("failed to lock user %s: %w", uq, err)
	}
	return nil
}

// unlockUser changes the user's shell back to /bin/bash to allow access
func (a *Actions) unlockUser(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	uq, err := safeUsername(username)
	if err != nil {
		return err
	}

	_, err = runCmd.Combined("dscl", ".", "-change", fmt.Sprintf("/Users/%s", uq), "UserShell", "/usr/bin/false", "/bin/bash")
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

	_, err = runCmd.Combined("dscl", ".", "-passwd", fmt.Sprintf("/Users/%s", uq), pq)
	if err != nil {
		return fmt.Errorf("failed to set password for user %s: %w", uq, err)
	}
	return nil
}

// addUser creates a new user, sets their password, and authorizes them to use FileVault
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

	// Create the user
	_, err = runCmd.Combined("dscl", ".", "-create", fmt.Sprintf("/Users/%s", uq))
	if err != nil {
		return fmt.Errorf("failed to create user %s: %w", uq, err)
	}

	// Set the user's password
	_, err = runCmd.Combined("dscl", ".", "-passwd", fmt.Sprintf("/Users/%s", uq), pq)
	if err != nil {
		return fmt.Errorf("failed to set password for user %s: %w", uq, err)
	}

	// Set the user's shell
	_, err = runCmd.Combined("dscl", ".", "-create", fmt.Sprintf("/Users/%s", uq), "UserShell", "/bin/bash")
	if err != nil {
		return fmt.Errorf("failed to set shell for user %s: %w", username, err)
	}

	// Set the user's home directory
	_, err = runCmd.Combined("dscl", ".", "-create", fmt.Sprintf("/Users/%s", uq), "NFSHomeDirectory", fmt.Sprintf("/Users/%s", username))
	if err != nil {
		return fmt.Errorf("failed to set home directory for user %s: %w", uq, err)
	}

	// Add the user to the list of users who can unlock the disk
	_, err = runCmd.Combined("fdesetup", "add", "-user", uq, "-password", pq)
	if err != nil {
		if strings.Contains(err.Error(), "FileVault is not enabled") {
			// Handle the case where FileVault is not enabled
			return nil
		} else {
			return fmt.Errorf("failed to add user %s to FileVault: %w", uq, err)
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
		_, err = runCmd.Combined("dseditgroup", "-o", "edit", "-a", uq, "-t", "user", "admin")
		if err != nil {
			return fmt.Errorf("failed to add user %s to admin group: %w", uq, err)
		}
	} else {
		_, err = runCmd.Combined("dseditgroup", "-o", "edit", "-d", uq, "-t", "user", "admin")
		if err != nil {
			return fmt.Errorf("failed to remove user %s from admin group: %w", uq, err)
		}
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

	// Attempt to remove from FileVault (best effort - may fail if FileVault not enabled or user not enrolled)
	_, err = runCmd.Combined("fdesetup", "remove", "-user", uq)
	if err != nil {
		// Log warning but continue - FileVault removal is not critical
		a.logger.Warningf(8409, "Failed to remove user %s from FileVault (may not be enrolled): %v", uq, err)
	}

	// Delete the user
	_, err = runCmd.Combined("dscl", ".", "-delete", fmt.Sprintf("/Users/%s", uq))
	if err != nil {
		return fmt.Errorf("failed to delete user %s: %w", uq, err)
	}

	return nil
}
