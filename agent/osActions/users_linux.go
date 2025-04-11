// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

// Code for Linux
//go:build linux

package osActions

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/UnifyEM/UnifyEM/common/schema"
)

// getUsers retrieves a list of all users on the system
func (a *Actions) getUsers() (schema.DeviceUserList, error) {
	// Get a list of users using getent
	cmd := exec.Command("getent", "passwd")
	output, err := cmd.Output()
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

// getSudoGroupMembers retrieves the members of the sudo/wheel group
func getSudoGroupMembers() (map[string]struct{}, error) {
	// Try sudo group first (Debian/Ubuntu)
	cmd := exec.Command("getent", "group", "sudo")
	output, err := cmd.Output()

	// If sudo group doesn't exist, try wheel (RHEL/CentOS/Fedora)
	if err != nil {
		cmd = exec.Command("getent", "group", "wheel")
		output, err = cmd.Output()
		if err != nil {
			return make(map[string]struct{}), nil // Return empty map if neither group exists
		}
	}

	// Parse the output (format: groupname:x:gid:user1,user2,...)
	parts := strings.Split(string(output), ":")
	if len(parts) < 4 {
		return make(map[string]struct{}), nil
	}

	members := strings.Split(parts[3], ",")
	sudoers := make(map[string]struct{}, len(members))
	for _, member := range members {
		member = strings.TrimSpace(member)
		if member != "" {
			sudoers[member] = struct{}{}
		}
	}
	return sudoers, nil
}

// isUserDisabled checks if the user has a valid shell
func isUserDisabled(shell string) bool {
	return shell == "/usr/sbin/nologin" || shell == "/bin/false" || shell == "/sbin/nologin"
}

// lockUser disables the user account by changing their shell to /usr/sbin/nologin
func (a *Actions) lockUser(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	uq, err := safeUsername(username)
	if err != nil {
		return err
	}

	// Remove quotes for usermod command
	uq = strings.Trim(uq, "\"")

	cmd := exec.Command("usermod", "-s", "/usr/sbin/nologin", uq)
	err = cmd.Run()
	if err != nil {
		// Try alternative location if /usr/sbin/nologin doesn't exist
		cmd = exec.Command("usermod", "-s", "/bin/false", uq)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to lock user %s: %w", uq, err)
		}
	}

	// Also lock the password
	cmd = exec.Command("passwd", "-l", uq)
	err = cmd.Run()
	if err != nil {
		a.logger.Errorf(8311, "Failed to lock password for user %s: %s", uq, err.Error())
		// Continue even if this fails, as changing the shell is the primary method
	}

	return nil
}

// unlockUser enables the user account by changing their shell back to /bin/bash
func (a *Actions) unlockUser(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	uq, err := safeUsername(username)
	if err != nil {
		return err
	}

	// Remove quotes for usermod command
	uq = strings.Trim(uq, "\"")

	cmd := exec.Command("usermod", "-s", "/bin/bash", uq)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to unlock user %s: %w", uq, err)
	}

	// Also unlock the password
	cmd = exec.Command("passwd", "-u", uq)
	err = cmd.Run()
	if err != nil {
		a.logger.Errorf(8312, "Failed to unlock password for user %s: %s", uq, err.Error())
		// Continue even if this fails, as changing the shell is the primary method
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

	// Remove quotes for the commands
	uq = strings.Trim(uq, "\"")
	pq = strings.Trim(pq, "\"")

	// Use chpasswd to set the password
	cmd := exec.Command("chpasswd")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s:%s", uq, pq))
	err = cmd.Run()
	if err != nil {
		// Fall back to passwd command if chpasswd fails
		a.logger.Errorf(8313, "chpasswd failed for user %s: %s", uq, err.Error())
		a.logger.Info(8314, "Falling back to passwd command", nil)

		// Create a temporary expect script to automate passwd
		tmpScript := fmt.Sprintf(`#!/usr/bin/expect -f
spawn passwd %s
expect "password:"
send "%s\r"
expect "password:"
send "%s\r"
expect eof
`, uq, pq, pq)

		// Write the script to a temporary file
		cmd = exec.Command("bash", "-c", fmt.Sprintf("cat > /tmp/passwd_script.exp << 'EOF'\n%s\nEOF", tmpScript))
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to create temporary script: %w", err)
		}

		// Make it executable
		cmd = exec.Command("chmod", "+x", "/tmp/passwd_script.exp")
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to make script executable: %w", err)
		}

		// Run the expect script
		cmd = exec.Command("/tmp/passwd_script.exp")
		err = cmd.Run()

		// Clean up
		cleanCmd := exec.Command("rm", "-f", "/tmp/passwd_script.exp")
		_ = cleanCmd.Run()

		if err != nil {
			return fmt.Errorf("failed to set password for user %s: %w", uq, err)
		}
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

	// Remove quotes for the commands
	uq = strings.Trim(uq, "\"")
	pq = strings.Trim(pq, "\"")

	// Create the user
	cmd := exec.Command("useradd", "-m", "-s", "/bin/bash", uq)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to create user %s: %w", uq, err)
	}

	// Set the user's password
	err = a.setPassword(uq, pq)
	if err != nil {
		return fmt.Errorf("failed to set password for user %s: %w", uq, err)
	}

	// Check if the user should be an admin
	if admin {
		return a.setAdmin(uq, true)
	}

	return nil
}

// setAdmin adds or removes a user from the admin group
func (a *Actions) setAdmin(username string, admin bool) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	uq, err := safeUsername(username)
	if err != nil {
		return err
	}

	// Remove quotes for usermod command
	uq = strings.Trim(uq, "\"")

	// Determine which admin group to use (sudo or wheel)
	adminGroup := "sudo" // Default for Debian/Ubuntu

	// Check if sudo group exists
	cmd := exec.Command("getent", "group", "sudo")
	err = cmd.Run()
	if err != nil {
		// Try wheel group (RHEL/CentOS/Fedora)
		cmd = exec.Command("getent", "group", "wheel")
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("neither sudo nor wheel group exists")
		}
		adminGroup = "wheel"
	}

	if admin {
		// Add user to admin group
		cmd = exec.Command("usermod", "-aG", adminGroup, uq)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to add user %s to %s group: %w", uq, adminGroup, err)
		}
	} else {
		// Remove user from admin group
		// Create a temporary script to remove user from group
		tmpScript := fmt.Sprintf(`#!/bin/bash
groups=\$(id -Gn %s | sed "s/ /,/g" | sed "s/%s//g" | sed "s/,,/,/g" | sed "s/^,//g" | sed "s/,$//g")
usermod -G "$groups" %s
`, uq, adminGroup, uq)

		// Write the script to a temporary file
		cmd = exec.Command("bash", "-c", fmt.Sprintf("cat > /tmp/remove_group.sh << 'EOF'\n%s\nEOF", tmpScript))
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to create temporary script: %w", err)
		}

		// Make it executable
		cmd = exec.Command("chmod", "+x", "/tmp/remove_group.sh")
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to make script executable: %w", err)
		}

		// Run the script
		cmd = exec.Command("/tmp/remove_group.sh")
		err = cmd.Run()

		// Clean up
		cleanCmd := exec.Command("rm", "-f", "/tmp/remove_group.sh")
		_ = cleanCmd.Run()

		if err != nil {
			return fmt.Errorf("failed to remove user %s from %s group: %w", uq, adminGroup, err)
		}
	}

	return nil
}
