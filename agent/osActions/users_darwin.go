//go:build darwin

/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package osActions

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/UnifyEM/UnifyEM/common/crypto"
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

	split := strings.Split(string(output), ":")
	if len(split) < 2 {
		return true
	}

	shell := strings.TrimSpace(split[1])
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

	_, err = runCmd.Combined("dscl", ".", "-create", fmt.Sprintf("/Users/%s", uq), "UserShell", "/usr/bin/false")
	if err != nil {
		return fmt.Errorf("failed to lock user %s: %w", uq, err)
	}
	return nil
}

// unlockUser changes the user's shell back to /bin/zsh to allow access
func (a *Actions) unlockUser(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	uq, err := safeUsername(username)
	if err != nil {
		return err
	}

	_, err = runCmd.Combined("dscl", ".", "-create", fmt.Sprintf("/Users/%s", uq), "UserShell", "/bin/zsh")
	if err != nil {
		return fmt.Errorf("failed to unlock user %s: %w", uq, err)
	}
	return nil
}

// setPassword sets the password for the specified user
func (a *Actions) setPassword(userInfo UserInfo) error {

	if userInfo.Username == "" || userInfo.Password == "" {
		return fmt.Errorf("username and password are required")
	}

	// Darwin also requires admin credentials to update FileVault
	err := a.TestCredentials(userInfo.AdminUser, userInfo.AdminPassword)
	if err != nil {
		return err
	}

	uq, err := safeUsername(userInfo.Username)
	if err != nil {
		return err
	}

	pq, err := safePassword(userInfo.Password)
	if err != nil {
		return err
	}

	// Remove the user from FileVault, otherwise we cannot set their password without
	// knowing the current one
	err = a.removeFileVault(uq)
	if err != nil {
		fmt.Printf("Failed to remove user %s from FileVault: %v", uq, err)
	}

	// Set the password
	_, err = runCmd.Combined("dscl", ".", "-passwd", fmt.Sprintf("/Users/%s", uq), pq)
	if err != nil {
		return fmt.Errorf("failed to set password for user %s: %w", uq, err)
	}

	// Add the user to FileVault with the new password
	return a.addFileVault(userInfo)
}

// addUser creates a new user, sets their password, and authorizes them to use FileVault
func (a *Actions) addUser(userInfo UserInfo) error {

	if userInfo.Username == "" || userInfo.Password == "" {
		return fmt.Errorf("username and password are required")
	}

	// Darwin also requires admin credentials to update FileVault
	if userInfo.AdminUser == "" || userInfo.AdminPassword == "" {
		return fmt.Errorf("administrator username and password must be supplied")
	}

	uq, err := safeUsername(userInfo.Username)
	if err != nil {
		return err
	}

	pq, err := safePassword(userInfo.Password)
	if err != nil {
		return err
	}

	// Set up the command
	var cmd = []string{"sysadminctl", "-addUser", uq, "-shell", "/bin/zsh", "-password", pq, "-home", fmt.Sprintf("/Users/%s", uq)}

	// Add admin if required
	if userInfo.Admin {
		cmd = append(cmd, "-admin")
	}

	//Create user
	_, err = runCmd.Combined(cmd...)
	if err != nil {
		return fmt.Errorf("failed to create user %s: %w", uq, err)
	}

	// Create home directory
	_, err = runCmd.Combined("createhomedir", "-c", "-u", uq)
	if err != nil {
		return fmt.Errorf("failed to create home directory for user %s: %w", uq, err)
	}

	// Update FileVault credentials
	return a.addFileVault(userInfo)
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
	_, err = runCmd.Combined("sysadminctl", "-deleteUser", uq)
	if err != nil {
		return fmt.Errorf("failed to deleteNo,  user %s: %w", uq, err)
	}

	return nil
}

func (a *Actions) addFileVault(userInfo UserInfo) error {

	// Darwin also requires admin credentials to update FileVault
	err := a.TestCredentials(userInfo.AdminUser, userInfo.AdminPassword)
	if err != nil {
		return err
	}

	// Validate required fields
	if userInfo.Username == "" || userInfo.Password == "" {
		return fmt.Errorf("username and password are required")
	}

	un, err := safeUsername(userInfo.Username)
	if err != nil {
		return err
	}

	up, err := safePassword(userInfo.Password)
	if err != nil {
		return err
	}

	an, err := safeUsername(userInfo.AdminUser)
	if err != nil {
		return err
	}

	ap, err := safePassword(userInfo.AdminPassword)
	if err != nil {
		return err
	}

	// Define the interactive prompts and responses
	interactive := runCmd.Interactive{

		Command: []string{"fdesetup", "add", "-usertoadd", un},
		Actions: []runCmd.Action{
			{
				WaitFor:  "Enter the user name:",
				Send:     an,
				DebugMsg: "Sending admin username",
			},
			{
				WaitFor:  "Enter the password for user",
				Send:     ap,
				DebugMsg: "Sending admin password",
			},
			{
				WaitFor:  "Enter the password for the added user",
				Send:     up,
				DebugMsg: "Sending user password",
			},
		},
	}

	// Run fdesetup with interactive prompts
	output, err := runCmd.TTY(interactive)
	if err != nil {
		// Check if FileVault is not enabled
		if strings.Contains(output, "FileVault is not enabled") {
			a.logger.Warningf(8410, "FileVault is not enabled, skipping FileVault configuration for user %s", un)
			return nil
		}

		// Check if user is already enabled for FileVault
		if strings.Contains(output, "already enabled") {
			a.logger.Infof(8411, "User %s is already enabled for FileVault", un)
			return nil
		}

		return fmt.Errorf("fdesetup failed: %w (output: %s)", err, output)
	}

	// Update the preboot
	_, err = runCmd.Combined("diskutil", "apfs", "updatePreboot", "/")
	if err != nil {
		// Log warning but don't fail - this is not critical
		a.logger.Warningf(8412, "Failed to update preboot after adding user %s to FileVault: %v", un, err)
	}

	return nil
}

// deleteUser removes a user from the system
func (a *Actions) removeFileVault(username string) error {
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
		return fmt.Errorf("failed to remove user %s from FileVault: %w", uq, err)
	}

	return nil
}

// userExists checks if a user exists on the system
func (a *Actions) userExists(username string) (bool, error) {

	if username == "" {
		return false, fmt.Errorf("username cannot be empty")
	}

	uq, err := safeUsername(username)
	if err != nil {
		return false, err
	}

	_, err = runCmd.Combined("dscl", ".", "-read", fmt.Sprintf("/Users/%s", uq))
	if err != nil {
		// Check if the error is because the user doesn't exist
		if strings.Contains(err.Error(), "eDSRecordNotFound") || strings.Contains(err.Error(), "No such file or directory") {
			return false, nil
		}
		// Some other error occurred
		return false, fmt.Errorf("failed to check if user %s exists: %w", uq, err)
	}

	return true, nil
}

// refreshServiceAccount changes the service account password using the old password for authentication
// Returns the new password on success
func (a *Actions) refreshServiceAccount(username, oldPassword string) (string, error) {
	if username == "" || oldPassword == "" {
		return "", fmt.Errorf("username and old password are required")
	}

	uq, err := safeUsername(username)
	if err != nil {
		return "", err
	}

	// Generate a new random password
	newPassword := crypto.RandomPassword()

	// Use dscl to change the password, authenticating with the old password
	// Format: dscl . -passwd /Users/<username> <oldpassword> <newpassword>
	_, err = runCmd.Combined("dscl", ".", "-passwd", fmt.Sprintf("/Users/%s", uq), oldPassword, newPassword)
	if err != nil {
		return "", fmt.Errorf("failed to change password for user %s: %w", uq, err)
	}

	return newPassword, nil
}

// testCredentials verifies that the username and password are valid,
// the user is an administrator, and the user has FileVault access (secure token)
func (a *Actions) testCredentials(username string, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("username and password are required")
	}

	uq, err := safeUsername(username)
	if err != nil {
		return err
	}

	pq, err := safePassword(password)
	if err != nil {
		return err
	}

	// Step 1: Verify the credentials are valid by attempting authentication with dscl
	// This uses the "authonly" option which authenticates without actually reading data
	_, err = runCmd.Combined("dscl", ".", "-authonly", uq, pq)
	if err != nil {
		return fmt.Errorf("authentication failed for user %s: invalid credentials", uq)
	}

	// Step 2: Verify the user is an administrator
	admins, err := getAdminGroupMembers()
	if err != nil {
		return fmt.Errorf("failed to check admin status: %w", err)
	}

	if _, isAdmin := admins[uq]; !isAdmin {
		return fmt.Errorf("user %s is not an administrator", uq)
	}

	// Step 3: Verify the user has a FileVault secure token
	output, err := runCmd.Combined("sysadminctl", "-secureTokenStatus", uq)
	if err != nil {
		return fmt.Errorf("failed to check secure token status for user %s: %w", uq, err)
	}

	outputStr := string(output)
	// The output format is typically: "Secure token is ENABLED for user <username>"
	if !strings.Contains(outputStr, "ENABLED") {
		return fmt.Errorf("user %s does not have a FileVault secure token", uq)
	}

	return nil
}
