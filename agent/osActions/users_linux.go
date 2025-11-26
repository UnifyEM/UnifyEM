//go:build linux

/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package osActions

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/UnifyEM/UnifyEM/common/crypto"
	"github.com/UnifyEM/UnifyEM/common/runCmd"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

// getUsers retrieves a list of all users on the system
func (a *Actions) getUsers() (schema.DeviceUserList, error) {

	// Get a list of users using getent
	output, err := runCmd.Stdout("getent", "passwd")
	if err != nil {
		return schema.DeviceUserList{}, fmt.Errorf("failed to get user list: %w", err)
	}

	// Get the sudo group members
	sudoers, err := getSudoGroupMembers()
	if err != nil {
		return schema.DeviceUserList{}, fmt.Errorf("failed to get admin group members: %w", err)
	}

	// Parse the output
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	var users schema.DeviceUserList
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 1 {
			continue
		}

		// Parse the passwd entry (format: username:x:uid:gid:gecos:home:shell)
		parts := strings.Split(line, ":")
		if len(parts) < 7 {
			continue
		}

		username := parts[0]
		uid, err := strconv.Atoi(parts[2])
		if err != nil {
			continue
		}

		// Skip system users (typically UID < 1000)
		if uid < 1000 && username != "root" {
			continue
		}

		// Check if the user's shell is a valid login shell
		shell := parts[6]
		disabled := isUserDisabled(shell)

		// Check if user is an admin (in sudo group)
		admin := false
		if _, isAdmin := sudoers[username]; isAdmin {
			admin = true
		}
		// Root is always an admin
		if username == "root" {
			admin = true
		}

		users.Users = append(users.Users,
			schema.DeviceUser{
				Name:          username,
				Domain:        "",
				Administrator: admin,
				Disabled:      disabled,
			},
		)
	}

	if err = scanner.Err(); err != nil {
		return schema.DeviceUserList{}, err
	}
	return users, nil
}

// getSudoGroupMembers retrieves the members of the sudo and wheel groups (union)
func getSudoGroupMembers() (map[string]struct{}, error) {
	sudoers := make(map[string]struct{})

	// Helper to parse group output
	parseMembers := func(output string) {
		parts := strings.Split(output, ":")
		if len(parts) >= 4 {
			for _, member := range strings.Split(parts[3], ",") {
				member = strings.TrimSpace(member)
				if member != "" {
					sudoers[member] = struct{}{}
				}
			}
		}
	}

	// Try getent group sudo
	output, err1 := runCmd.Stdout("getent", "group", "sudo")
	if err1 == nil && len(output) > 0 {
		parseMembers(output)
	}

	// Try getent group wheel
	output, err2 := runCmd.Stdout("getent", "group", "wheel")
	if err2 == nil && len(output) > 0 {
		parseMembers(output)
	}

	// If both commands failed to execute (not just empty output), treat as error
	if (err1 != nil && err2 != nil) && len(sudoers) == 0 {
		// Only return empty map if both commands failed to execute (not just empty output)
		return make(map[string]struct{}), nil
	}

	return sudoers, nil
}

// isUserDisabled checks if the user has a valid shell
func isUserDisabled(shell string) bool {
	return shell == "/usr/sbin/nologin" || shell == "/bin/false" || shell == "/sbin/nologin"
}

// lockUser disables the user account by changing their shell to /usr/sbin/nologin
func (a *Actions) lockUser(userInfo UserInfo) error {
	if userInfo.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	_, err := runCmd.Combined("usermod", "-s", "/usr/sbin/nologin", userInfo.Username)
	if err != nil {
		// Try alternative location if /usr/sbin/nologin doesn't exist
		_, err = runCmd.Combined("usermod", "-s", "/bin/false", userInfo.Username)
		if err != nil {
			return fmt.Errorf("failed to lock user %s: %w", userInfo.Username, err)
		}
	}

	// Also lock the password
	_, err = runCmd.Combined("passwd", "-l", userInfo.Username)
	if err != nil {
		a.logger.Errorf(8311, "Failed to lock password for user %s: %s", userInfo.Username, err.Error())
		// Continue even if this fails, as changing the shell is the primary method
	}

	return nil
}

// unlockUser enables the user account by changing their shell back to a valid bash shell
func (a *Actions) unlockUser(userInfo UserInfo) error {
	if userInfo.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Determine which shell exists (prefer bash, fall back to sh)
	shell := "/bin/bash"
	if _, statErr := os.Stat("/bin/bash"); statErr != nil {
		if _, statErr2 := os.Stat("/usr/bin/bash"); statErr2 == nil {
			shell = "/usr/bin/bash"
		} else if _, statErr3 := os.Stat("/bin/sh"); statErr3 == nil {
			shell = "/bin/sh"
		} else {
			return fmt.Errorf("no valid shell found for user %s (tried /bin/bash, /usr/bin/bash, /bin/sh)", userInfo.Username)
		}
	}

	_, err := runCmd.Combined("usermod", "-s", shell, userInfo.Username)
	if err != nil {
		return fmt.Errorf("failed to unlock user %s: %w", userInfo.Username, err)
	}

	// Also unlock the password
	_, err = runCmd.Combined("passwd", "-u", userInfo.Username)
	if err != nil {
		a.logger.Errorf(8312, "Failed to unlock password for user %s: %s", userInfo.Username, err.Error())
		// Continue even if this fails, as changing the shell is the primary method
	}

	return nil
}

// setPassword sets the password for the specified user
func (a *Actions) setPassword(userInfo UserInfo) error {

	if userInfo.Username == "" || userInfo.Password == "" {
		return fmt.Errorf("username and password are required")
	}

	// Use chpasswd to set the password
	cmd := exec.Command("chpasswd")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s:%s", userInfo.Username, userInfo.Password))
	err := cmd.Run()
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

	// Create the user
	_, err := runCmd.Combined("useradd", "-m", "-s", "/bin/bash", userInfo.Username)
	if err != nil {
		return fmt.Errorf("failed to create user %s: %w", userInfo.Username, err)
	}

	// Set the user's password
	err = a.setPassword(userInfo)
	if err != nil {
		return fmt.Errorf("failed to set password for user %s: %w", userInfo.Username, err)
	}

	// Check if the user should be an admin
	if userInfo.Admin {
		return a.setAdmin(userInfo)
	}

	return nil
}

// setAdmin adds or removes a user from the admin group
func (a *Actions) setAdmin(userInfo UserInfo) error {
	if userInfo.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Determine which admin group to use (sudo or wheel)
	adminGroup := "sudo" // Default for Debian/Ubuntu

	// Check if sudo group exists
	_, err := runCmd.Combined("getent", "group", "sudo")
	if err != nil {
		// Try wheel group (RHEL/CentOS/Fedora)
		_, err = runCmd.Combined("getent", "group", "wheel")
		if err != nil {
			return fmt.Errorf("neither sudo nor wheel group exists")
		}
		adminGroup = "wheel"
	}

	if userInfo.Admin {
		// Add user to admin group
		_, err := runCmd.Combined("usermod", "-aG", adminGroup, userInfo.Username)
		if err != nil {
			return fmt.Errorf("failed to add user %s to %s group: %w", userInfo.Username, adminGroup, err)
		}

		// Check if /etc/sudoers.d exists
		if stat, statErr := os.Stat("/etc/sudoers.d"); statErr == nil && stat.IsDir() {
			sudoersFile := fmt.Sprintf("/etc/sudoers.d/%s-UEM", userInfo.Username)
			content := fmt.Sprintf("%s ALL=(ALL) ALL\n", userInfo.Username)
			writeErr := os.WriteFile(sudoersFile, []byte(content), 0440)
			if writeErr != nil {
				return fmt.Errorf("failed to create sudoers file for user %s: %w", userInfo.Username, writeErr)
			}
			chmodErr := os.Chmod(sudoersFile, 0440)
			if chmodErr != nil {
				return fmt.Errorf("failed to set permissions on sudoers file for user %s: %w", userInfo.Username, chmodErr)
			}
		}
	} else {
		// Remove user from admin group using gpasswd
		_, err = runCmd.Combined("gpasswd", "-d", userInfo.Username, adminGroup)
		if err != nil {
			return fmt.Errorf("failed to remove user %s from %s group: %w", userInfo.Username, adminGroup, err)
		}

		// Remove sudoers file if it exists
		sudoersFile := fmt.Sprintf("/etc/sudoers.d/%s-UEM", userInfo.Username)
		if _, statErr := os.Stat(sudoersFile); statErr == nil {
			removeErr := os.Remove(sudoersFile)
			if removeErr != nil {
				return fmt.Errorf("failed to remove sudoers file for user %s: %w", userInfo.Username, removeErr)
			}
		}
	}
	return nil
}

// deleteUser removes a user from the system
func (a *Actions) deleteUser(userInfo UserInfo) error {
	if userInfo.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Remove sudoers file if it exists
	sudoersFile := fmt.Sprintf("/etc/sudoers.d/%s-UEM", userInfo.Username)
	if _, statErr := os.Stat(sudoersFile); statErr == nil {
		removeErr := os.Remove(sudoersFile)
		if removeErr != nil {
			a.logger.Warningf(8313, "Failed to remove sudoers file for user %s: %v", userInfo.Username, removeErr)
			// Continue even if this fails
		}
	}

	// Delete the user
	_, err := runCmd.Combined("userdel", userInfo.Username)
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

	_, err := runCmd.Stdout("getent", "passwd", username)
	if err != nil {
		// Check if the error is because the user doesn't exist
		if strings.Contains(err.Error(), "exit status 2") {
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

	// Use su to validate credentials
	// Create a command that attempts to switch to the user with their password
	cmd := exec.Command("su", "-c", "exit 0", user)
	cmd.Stdin = strings.NewReader(pass + "\n")

	// Capture output to check for specific error messages
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check for authentication failure vs other errors
		outputStr := string(output)
		if strings.Contains(outputStr, "Authentication failure") ||
			strings.Contains(outputStr, "incorrect password") {
			return fmt.Errorf("authentication failed for user %s: invalid credentials", user)
		}
		return fmt.Errorf("failed to validate credentials for user %s: %w", user, err)
	}

	return nil
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

	// Set the new password using chpasswd (runs as root, no old password needed)
	cmd := exec.Command("chpasswd")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s:%s", userInfo.Username, newPassword))
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to change password for user %s: %w", userInfo.Username, err)
	}

	return newPassword, nil
}
