//go:build darwin

/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package osActions

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/UnifyEM/UnifyEM/common"
	"github.com/UnifyEM/UnifyEM/common/crypto"
	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/runCmd"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

func (a *Actions) getUsers() (schema.DeviceUserList, error) {

	// Get a list of users
	a.logger.Debugf(8431, "Calling dscl to obtain list of users")
	output, err := a.runner.Stdout("dscl", ".", "list", "/Users")
	if err != nil {
		return schema.DeviceUserList{}, err
	}

	// Get the admin group members
	admins, err := a.getAdminGroupMembers()
	if err != nil {
		return schema.DeviceUserList{}, err
	}

	// Parse the output
	scanner := bufio.NewScanner(strings.NewReader(output))
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
				Disabled:      !a.canUserLogin(user),
			},
		)
	}

	if err = scanner.Err(); err != nil {
		return schema.DeviceUserList{}, err
	}

	return users, nil
}

// getAdminGroupMembers retrieves the members of the admin group
func (a *Actions) getAdminGroupMembers() (map[string]struct{}, error) {
	a.logger.Debugf(8432, "Calling dscl to obtain list of users")
	output, err := a.runner.Stdout("dscl", ".", "read", "/Groups/admin", "GroupMembership")
	if err != nil {
		return nil, err
	}

	members := strings.Fields(output)
	admins := make(map[string]struct{}, len(members))
	for _, member := range members {
		admins[member] = struct{}{}
	}
	return admins, nil
}

// canUserLogin checks if the user has a valid shell
func (a *Actions) canUserLogin(username string) bool {
	if username == "" {
		return false
	}

	a.logger.Debugf(8433, "Calling dscl to obtain information for user %s", username)
	output, err := a.runner.Stdout("dscl", ".", "-read", "/Users/"+username, "UserShell")
	if err != nil {
		return false
	}

	split := strings.Split(output, ":")
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
func (a *Actions) lockUser(userInfo UserInfo) error {
	var err error

	if userInfo.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Darwin also requires admin credentials to update FileVault
	err = a.TestCredentials(userInfo.AdminUser, userInfo.AdminPassword)
	if err != nil {
		return err
	}

	// Remove the user's Secure Token
	// This also changes their password to a random one
	_, err = a.removeSecureToken(userInfo)
	if err != nil {
		a.logger.Warningf(8461, "error removing secure token: %s", common.SingleLine(err.Error()))
		// Continue anyway
	}

	// Remove the user from FileVault
	err = a.removeFileVault(userInfo)
	if err != nil {
		a.logger.Warningf(8417, "Failed to remove user %s from FileVault (may not be enrolled): %v", userInfo.Username, err)
	}

	a.logger.Debugf(8434, "Calling dscl to set shell for user %s to /usr/bin/false", userInfo.Username)
	_, err = a.runner.Combined("dscl", ".", "-create", fmt.Sprintf("/Users/%s", userInfo.Username), "UserShell", "/usr/bin/false")
	if err != nil {
		return fmt.Errorf("failed to lock user %s: %w", userInfo.Username, err)
	}
	return nil
}

// unlockUser changes the user's shell back to /bin/zsh to allow access
func (a *Actions) unlockUser(userInfo UserInfo) error {

	if userInfo.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Darwin also requires admin credentials to update FileVault
	err := a.TestCredentials(userInfo.AdminUser, userInfo.AdminPassword)
	if err != nil {
		return err
	}

	// Enable user
	a.logger.Debugf(8435, "Calling dscl to set shell for user %s to /bin/zsh", userInfo.Username)
	_, err = a.runner.Combined("dscl", ".", "-create", fmt.Sprintf("/Users/%s", userInfo.Username), "UserShell", "/bin/zsh")
	if err != nil {
		return fmt.Errorf("failed to unlock user %s: %w", userInfo.Username, err)
	}

	// Set the password and add back to fileVault
	return a.setPassword(userInfo)
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

	// Remove the user from FileVault, otherwise we cannot set their password without
	// knowing the current one
	err = a.removeFileVault(userInfo)
	if err != nil {
		a.logger.Warningf(8415, "Failed to remove user %s from FileVault (may not be enrolled): %v", userInfo.Username, err)
		// Don't return - if they're in FV and removal failed, password change will fail below
	}

	// Set the password
	a.logger.Debugf(8435, "Calling dscl to set password for user %s", userInfo.Username)
	_, err = a.runner.Combined("dscl", ".", "-passwd", fmt.Sprintf("/Users/%s", userInfo.Username), userInfo.Password)
	if err != nil {
		return fmt.Errorf("failed to set password for user %s: %w", userInfo.Username, err)
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

	// Set up the command
	a.logger.Debugf(8436, "Calling sysadminctl to add user %s", userInfo.Username)
	var cmd = []string{"sysadminctl", "-addUser", userInfo.Username, "-shell", "/bin/zsh", "-password", userInfo.Password, "-home", fmt.Sprintf("/Users/%s", userInfo.Username)}

	// Add admin if required
	if userInfo.Admin {
		cmd = append(cmd, "-admin")
	}

	// Create user
	_, err := a.runner.Combined(cmd...)
	if err != nil {
		return fmt.Errorf("failed to create user %s: %w", userInfo.Username, err)
	}

	// Create home directory
	a.logger.Debugf(8437, "Calling createhomedir to create home directory for %s", userInfo.Username)
	_, err = a.runner.Combined("createhomedir", "-c", "-u", userInfo.Username)
	if err != nil {
		return fmt.Errorf("failed to create home directory for user %s: %w", userInfo.Username, err)
	}

	// Update FileVault credentials
	return a.addFileVault(userInfo)
}

func (a *Actions) setAdmin(userInfo UserInfo) error {

	if userInfo.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if userInfo.Admin {
		a.logger.Debugf(8438, "Calling dseditgroup to add user %s to admin group", userInfo.Username)
		_, err := a.runner.Combined("dseditgroup", "-o", "edit", "-a", userInfo.Username, "-t", "user", "admin")
		if err != nil {
			return fmt.Errorf("failed to add user %s to admin group: %w", userInfo.Username, err)
		}
	} else {
		a.logger.Debugf(8439, "Calling dseditgroup to remove user %s from admin group", userInfo.Username)
		_, err := a.runner.Combined("dseditgroup", "-o", "edit", "-d", userInfo.Username, "-t", "user", "admin")
		if err != nil {
			return fmt.Errorf("failed to remove user %s from admin group: %w", userInfo.Username, err)
		}
	}
	return nil
}

// deleteUser removes a user from the system
//
// macOS only partially deletes the user and does delete their home directory
// Lock the user instead
func (a *Actions) deleteUser(userInfo UserInfo) error {
	return a.lockUser(userInfo)

	/*

		var err, exitErr error
		var success bool

		if userInfo.Username == "" {
			return fmt.Errorf("username cannot be empty")
		}

		// Darwin also requires admin credentials to update FileVault
		err = a.TestCredentials(userInfo.AdminUser, userInfo.AdminPassword)
		if err != nil {
			return err
		}

		// Assume error
		exitErr = fmt.Errorf("error deleting user %s", userInfo.Username)
		success = true

		_, err = a.removeSecureToken(userInfo)
		if err != nil {
			a.logger.Warningf(8461, "error removing secure token: %s", common.SingleLine(err.Error()))
			success = false
		}

		_, err = a.deleteUserWithSSH(userInfo)
		if err != nil {
			a.logger.Warningf(8468, "error deleting user with SSH: %s", common.SingleLine(err.Error()))
			success = false
		}

			_, err = a.deleteUserWithLC(userInfo)
			if err != nil {
				a.logger.Warningf(8462, "error deleting user with launchctl: %s", common.SingleLine(err.Error()))
				success = false
			}

			_, err = a.deleteUserWithAdmin(userInfo)
			if err != nil {
				a.logger.Infof(8463, "sysadminctl error deleting user %s: %s", userInfo.Username, common.SingleLine(err.Error()))
				a.logger.Infof(8464, "attempting to delete user %s with dscl", userInfo.Username)

				// Try with dscl
				_, err = a.deleteUserWithDscl(userInfo)
				if err != nil {
					a.logger.Infof(8465, "sysadminctl error deleting user %s: %s", userInfo.Username, common.SingleLine(err.Error()))
					success = false
				}
			}

		// Remove from FileVault just in case
		err = a.removeFileVault(userInfo)
		if err != nil {
			a.logger.Infof(8466, "failed to remove user %s from FileVault: %s", userInfo.Username, common.SingleLine(err.Error()))
			success = false
		}

		if success {
			return nil
		}
		return exitErr
	*/
}

func (a *Actions) deleteUserWithAdmin(userInfo UserInfo) (string, error) {

	// macOS also requires admin credentials to update FileVault
	if userInfo.AdminUser == "" || userInfo.AdminPassword == "" {
		return "", fmt.Errorf("administrator username and password must be supplied")
	}

	a.logger.Debugf(8440, "Calling sysadminctl (as service account) to delete user %s", userInfo.Username)

	// Escape the sandbox and run sysadminctl as root via sudo
	out, err := a.runner.TTYAsUser(
		&runCmd.UserLogin{
			Username:  userInfo.AdminUser,
			Password:  userInfo.AdminPassword,
			RunAsRoot: true,
		},
		"sysadminctl", "-deleteUser", userInfo.Username, "-adminUser", userInfo.AdminUser, "-adminPassword", userInfo.AdminPassword)

	//out, err := a.runner.Combined("sysadminctl", "-deleteUser", userInfo.Username, "-adminUser", userInfo.AdminUser, "-adminPassword", userInfo.AdminPassword)
	if err != nil {
		return common.SingleLine(out), fmt.Errorf("failed to delete user %s: %w", userInfo.Username, err)
	}

	if strings.Contains(out, "-14120") || strings.Contains(out, "Error:") {
		return common.SingleLine(out), fmt.Errorf("failed to delete user %s: %s", userInfo.Username, common.SingleLine(out))
	}

	return common.SingleLine(out), nil
}

func (a *Actions) deleteUserWithDscl(userInfo UserInfo) (string, error) {

	if userInfo.Username == "" {
		return "", fmt.Errorf("username cannot be empty")
	}

	a.logger.Debugf(8441, "Calling dscl (as service account) to delete user %s", userInfo.Username)

	// Escape the sandbox and run dscl as root via sudo
	out, err := a.runner.TTYAsUser(
		&runCmd.UserLogin{
			Username:  userInfo.AdminUser,
			Password:  userInfo.AdminPassword,
			RunAsRoot: true,
		},
		"dscl", ".", "-delete", fmt.Sprintf("/Users/%s", userInfo.Username))

	//out, err := a.runner.Combined("dscl", ".", "-delete", fmt.Sprintf("/Users/%s", userInfo.Username))
	if err != nil {
		return common.SingleLine(out), fmt.Errorf("failed to delete user %s: %w", userInfo.Username, err)
	}
	return common.SingleLine(out), nil
}
func (a *Actions) deleteUserWithSSH(userInfo UserInfo) (string, error) {
	if userInfo.AdminUser == "" || userInfo.AdminPassword == "" {
		return "", fmt.Errorf("administrator username and password must be supplied")
	}

	a.logger.Debugf(8467, "Calling sysadminctl via SSH as service account to delete user %s", userInfo.Username)

	// Use SSH to localhost to escape the sandbox and run sysadminctl as root
	out, err := a.runner.SSH(
		&runCmd.UserLogin{
			Username:  userInfo.AdminUser,
			Password:  userInfo.AdminPassword,
			RunAsRoot: true,
		},
		"sysadminctl", "-deleteUser", userInfo.Username,
		"-adminUser", userInfo.AdminUser,
		"-adminPassword", userInfo.AdminPassword)

	if err != nil {
		return common.SingleLine(out), fmt.Errorf("failed to SSH DELETE user %s: %w", userInfo.Username, err)
	}

	if strings.Contains(out, "-14120") || strings.Contains(out, "Error:") {
		return common.SingleLine(out), fmt.Errorf("failed to delete user %s: %s", userInfo.Username, common.SingleLine(out))
	}

	return common.SingleLine(out), nil
}

func (a *Actions) deleteUserWithLC(userInfo UserInfo) (string, error) {

	// Set up the command
	var cmd = []string{"launchctl", "asuser", "504", "sysadminctl", "-deleteUser", userInfo.Username, "-adminUser", userInfo.AdminUser, "-adminPassword", userInfo.AdminPassword}

	a.logger.Debugf(8442, "Calling sysadminctl via launchctl (as service account) to delete user %s", userInfo.Username)

	out, err := a.runner.Combined(cmd...)
	fmt.Printf("\n\n*****\n%s\n*****\n\n", out)
	if err != nil {
		return out, fmt.Errorf("failed to LC DELETE user %s: %w", userInfo.Username, err) // TODO
	}
	return out, nil
}

func (a *Actions) addFileVault(userInfo UserInfo) error {

	// Validate required fields
	if userInfo.Username == "" || userInfo.Password == "" {
		return fmt.Errorf("username and password are required")
	}

	// Darwin also requires admin credentials to update FileVault
	err := a.TestCredentials(userInfo.AdminUser, userInfo.AdminPassword)
	if err != nil {
		return err
	}

	// Define the interactive prompts and responses
	interactive := runCmd.Interactive{

		Command: []string{"sysadminctl", "-adminUser", userInfo.AdminUser, "-adminPassword", "-",
			"-secureTokenOn", userInfo.Username, "-password", "-"},
		Actions: []runCmd.Action{
			{
				WaitFor:  "Enter password for ",
				Send:     userInfo.AdminPassword,
				DebugMsg: "Sending admin password",
			},
			{
				WaitFor:  "Enter password for ",
				Send:     userInfo.Password,
				DebugMsg: "Sending user password",
			},
		},
	}

	// Log attempt to add user to FileVault
	a.logger.Info(8443, "Calling sysadminctl with service account credentials to grant secure token",
		fields.NewFields(fields.NewField("user", userInfo.Username)))

	// Run fdesetup with interactive prompts
	output, err := a.runner.TTY(interactive)
	if err != nil {
		// Check if FileVault is not enabled
		if strings.Contains(output, "FileVault is not enabled") {
			a.logger.Warningf(8410, "FileVault is not enabled, skipping FileVault configuration for user %s", userInfo.Username)
			return nil
		}

		// Check if user is already enabled for FileVault
		if strings.Contains(output, "already enabled") {
			a.logger.Infof(8411, "User %s is already enabled for FileVault", userInfo.Username)
			return nil
		}

		// Log the failure with output for debugging
		a.logger.Errorf(8423, "granting secure token failed for user %s: %v",
			userInfo.Username, err,
			fields.NewFields(fields.NewField("output", output)))
		return fmt.Errorf("granting secure token failed: %w (output: %s)", err, output)
	}

	// Log successful fdesetup execution
	a.logger.Info(8424, "secure token granted successfully",
		fields.NewFields(fields.NewField("user", userInfo.Username)))

	a.logger.Debugf(8444, "Calling diskutil to update preboot after adding user %s to FileVault", userInfo.Username)

	// Update the preboot
	_, err = a.runner.Combined("diskutil", "apfs", "updatePreboot", "/")
	if err != nil {
		// Log warning but don't fail - this is not critical
		a.logger.Warningf(8412, "Failed to update preboot after adding user %s to FileVault: %v", userInfo.Username, err)
	}

	a.logger.Info(8417, "user successfully added to FileVault with secure token",
		fields.NewFields(fields.NewField("user", userInfo.Username)))
	return nil
}

// removeFileVault removes a user from FileVault
func (a *Actions) removeFileVault(userInfo UserInfo) error {

	if userInfo.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	a.logger.Debugf(8445, "Calling fdesetup to remove user %s from FileVault", userInfo.Username)

	// Attempt to remove from FileVault
	_, err := a.runner.Combined("fdesetup", "remove", "-user", userInfo.Username)
	if err != nil {
		if strings.Contains(err.Error(), "User could not be found") {
			return nil
		}
		return fmt.Errorf("failed to remove user %s from FileVault: %w", userInfo.Username, err)
	}
	return nil
}

// removeSecureToken changes a user's password to a random value and removes their secure token
func (a *Actions) removeSecureToken(userInfo UserInfo) (string, error) {
	fmt.Printf("removeSecureToken called\n")

	if userInfo.Username == "" {
		return "", fmt.Errorf("username cannot be empty")
	}

	// Darwin also requires admin credentials to update FileVault
	err := a.TestCredentials(userInfo.AdminUser, userInfo.AdminPassword)
	if err != nil {
		return "", err
	}

	// Generate a temporary random password for the user
	// This is needed because sysadminctl requires the user's password to remove the secure token
	tempPassword := crypto.RandomPassword()
	a.logger.Debugf(8446, "Setting temporary password for user %s to enable removing secure token", userInfo.Username)

	// Change the user's password to our known temporary value
	// Use passwd command which can be run by root/admin without knowing current password
	err = a.setPassword(UserInfo{
		Username:      userInfo.Username,
		Password:      tempPassword,
		AdminUser:     userInfo.AdminUser,
		AdminPassword: userInfo.AdminPassword})
	if err != nil {
		a.logger.Warningf(8462, "Failed to set temporary password for user %s: %v", userInfo.Username, err)
		// Continue anyway - secure token removal might still work
	}

	a.logger.Debugf(8446, "Calling sysadminctl to remove secure token for user %s", userInfo.Username)

	// Attempt to remove the secure token using the temporary password
	out, err := a.runner.Combined("sysadminctl", "-secureTokenOff", userInfo.Username, "-adminUser",
		userInfo.AdminUser, "-adminPassword", userInfo.AdminPassword, "-password", tempPassword)
	if err != nil {
		if strings.Contains(err.Error(), "User could not be found") {
			return out, nil
		}
		return out, fmt.Errorf("failed to remove secure token for user %s: %w", userInfo.Username, err)
	}
	return out, nil
}

// userExists checks if a user exists on the system
func (a *Actions) userExists(username string) (bool, error) {

	if username == "" {
		return false, fmt.Errorf("username cannot be empty")
	}

	a.logger.Debugf(8447, "Calling dscl to determine if user %s exists or not", username)

	_, err := a.runner.Combined("dscl", ".", "-read", fmt.Sprintf("/Users/%s", username))
	if err != nil {
		// Check if the error is because the user doesn't exist
		if strings.Contains(err.Error(), "eDSRecordNotFound") || strings.Contains(err.Error(), "No such file or directory") {
			return false, nil
		}
		// Some other error occurred
		return false, fmt.Errorf("failed to check if user %s exists: %w", username, err)
	}

	return true, nil
}

// refreshServiceAccount changes the service account password using the old password for authentication
// and ensures the account is an administrator. Returns the new password on success
func (a *Actions) refreshServiceAccount(userInfo UserInfo) (string, error) {
	if userInfo.Username == "" || userInfo.Password == "" {
		return "", fmt.Errorf("username and existing password are required")
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

	a.logger.Debugf(8448, "Calling dscl to change service account (%s) password", userInfo.Username)

	// Use dscl to change the password, authenticating with the old password
	// Format: dscl . -passwd /Users/<username> <oldpassword> <newpassword>
	_, err = a.runner.Combined("dscl", ".", "-passwd", fmt.Sprintf("/Users/%s", userInfo.Username), userInfo.Password, newPassword)
	if err != nil {
		return "", fmt.Errorf("failed to change password for user %s: %w", userInfo.Username, err)
	}

	return newPassword, nil
}

// testCredentials verifies that the username and password are valid,
// the user is an administrator, and the user has FileVault access (secure token)
func (a *Actions) testCredentials(username string, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("credential test failed, username and password cannot be empty")
	}

	// Verify the credentials are valid by attempting authentication with dscl
	// This uses the "authonly" option which authenticates without actually reading data

	a.logger.Debugf(8449, "Calling dscl with -authonly to test credentials for user %s password", username)

	_, err := a.runner.Combined("dscl", ".", "-authonly", username, password)
	if err != nil {
		return fmt.Errorf("authentication failed for user %s: invalid credentials", username)
	}

	// Verify the user is an administrator
	admins, err := a.getAdminGroupMembers()
	if err != nil {
		return fmt.Errorf("failed to check admin status: %w", err)
	}

	if _, isAdmin := admins[username]; !isAdmin {
		return fmt.Errorf("user %s is not an administrator", username)
	}

	// Verify the user has a secure token (for filevault)
	a.logger.Debugf(8450, "Calling sysadminctl to check secure token status for user %s", username)
	out, err := a.runner.Combined("sysadminctl", "-secureTokenStatus", username)
	if err != nil {
		return fmt.Errorf("failed to check secure token status for user %s: %w", username, err)
	}

	// The output format is typically: "Secure token is ENABLED for user <username>"
	if !strings.Contains(out, "ENABLED") {
		return fmt.Errorf("user %s does not have a FileVault secure token", username)
	}

	return nil
}
