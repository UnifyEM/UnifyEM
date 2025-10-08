//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

// MacOS (Darin) specific functions
//go:build darwin

package install

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/UnifyEM/UnifyEM/agent/global"
)

const (
	serviceName     = "uem-agent"
	binaryPath      = "/usr/local/bin"
	daemonPlistPath = "/Library/LaunchDaemons/com.tenebris.uem-agent.plist"
	agentPlistPath  = "/Library/LaunchAgents/com.tenebris.uem-agent.plist"
)

// Note that this must also be changed if binaryPath or global.UnixBinaryName are changed
//
//goland:noinspection HttpUrlsUsage
const daemonPlistContent = `
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.tenebris.uem-agent</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/uem-agent</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>StartInterval</key>
    <integer>60</integer>
    <key>UserName</key>
    <string>root</string>
    <key>StandardOutPath</key>
    <string>/var/log/uem-agent.log</string>
    <key>StandardErrorPath</key>
    <string>/var/log/uem-agent.log</string>
</dict>
</plist>
`

//goland:noinspection HttpUrlsUsage
const agentPlistContent = `
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.tenebris.uem-agent-user</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/uem-agent</string>
        <string>--user-helper</string>
        <string>--collection-interval</string>
        <string>300</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/uem-agent-user.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/uem-agent-user.log</string>
</dict>
</plist>
`


// Install the service
func (i *Install) installService() error {

	// Get the path of the current executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find executable path: %w", err)
	}

	// Set the target path
	targetPath := binaryPath + string(os.PathSeparator) + serviceName

	// Copy the executable to the target directory
	if exePath != targetPath {
		err = copyFile(exePath, targetPath)
		if err != nil {
			return fmt.Errorf("error copying file %s to %s: %v", exePath, targetPath, err)
		}
	}
	fmt.Printf("Binary copied to %s\n", targetPath)

	// Set the proper permissions on the binary (755 allows user-helper mode to run as non-root)
	err = os.Chmod(targetPath, 0755)
	if err != nil {
		return fmt.Errorf("could not set permissions on binary: %w", err)
	}

	// Set the owner of the binary to root
	cmd := exec.Command("chown", "root:wheel", targetPath)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("could not set owner of binary to root: %w", err)
	}

	// Create daemon plist
	err = i.createPlist(daemonPlistPath, daemonPlistContent)
	if err != nil {
		return err
	}

	// Create agent plist
	err = i.createPlist(agentPlistPath, agentPlistContent)
	if err != nil {
		return err
	}

	// Load the Launch Daemon
	cmd = exec.Command("launchctl", "load", daemonPlistPath)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("could not load launch daemon: %w", err)
	}

	// Bootstrap user agent for currently logged-in users
	// Note: The plist in /Library/LaunchAgents/ will auto-load for users at login
	// We only need to bootstrap for currently logged-in users
	bootstrapUserAgents()

	return nil
}

// Uninstall the service
func (i *Install) uninstallService(removeData bool) error {

	// Set the target path
	targetPath := binaryPath + string(os.PathSeparator) + serviceName

	// Unload the Launch Daemon
	cmd := exec.Command("launchctl", "unload", daemonPlistPath)
	err := cmd.Run()
	if err != nil {
		// Delay 30 seconds and try again
		time.Sleep(30 * time.Second)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("could not unload launch daemon: %w", err)
		}
	}

	// Unload all user agent instances
	unloadUserAgents()

	// Wait for the program to terminate
	time.Sleep(5 * time.Second)

	// Check if the process is still running
	for i := 0; i < 36; i++ { // Wait up to 3 minutes
		if !isProcessRunning(global.UnixBinaryName) {
			break
		}
		time.Sleep(5 * time.Second)
	}

	if isProcessRunning(global.UnixBinaryName) {
		return fmt.Errorf("program did not terminate within the expected time")
	}

	// Remove daemon plist
	err = os.Remove(daemonPlistPath)
	if err != nil {
		return fmt.Errorf("could not remove daemon plist file: %w", err)
	}

	// Remove agent plist
	err = os.Remove(agentPlistPath)
	if err != nil {
		fmt.Printf("Warning: could not remove agent plist: %v\n", err)
	}

	// Remove the service binary
	err = os.Remove(targetPath)
	if err != nil {
		return fmt.Errorf("could not remove binary file: %w", err)
	}

	if removeData {
		// TODO delete the data
	}

	return nil
}

// isProcessRunning checks if a process with the given name is running
func isProcessRunning(name string) bool {
	cmd := exec.Command("pgrep", name)
	err := cmd.Run()
	return err == nil
}

// Upgrade the service
func (i *Install) upgradeService() error {

	fmt.Println("Uninstalling existing agent...")

	// Remove the existing executable
	err := i.uninstallService(false)
	if err != nil {
		return fmt.Errorf("could not remove existing service: %w", err)
	}

	// Delay for two seconds to allow the system to release the file
	time.Sleep(2 * time.Second)

	fmt.Println("\nInstalling new agent...")

	// Install the new service
	return i.installService()
}

// CheckRootPrivileges checks if the current user has root privileges and if not,
// it will attempt to gain root privileges by running the current program with sudo
func CheckRootPrivileges() error {
	if os.Geteuid() != 0 {
		fmt.Println("\nThis program must be run as root, restarting with sudo...")
		cmd := exec.Command("sudo", os.Args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to gain root privileges: %w", err)
		}
		os.Exit(0)
	}
	return nil
}

// createPlist creates a plist file at the specified path with the given content
func (i *Install) createPlist(path string, content string) error {
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("could not write plist file: %w", err)
	}

	fmt.Printf("Plist file created at: %s\n", path)
	return nil
}

// stopService stops the service
func (i *Install) stopService() error {
	cmd := exec.Command("launchctl", "stop", daemonPlistPath)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error stopping service: %w", err)
	}
	return nil
}

// startService starts the service
func (i *Install) startService() error {
	cmd := exec.Command("launchctl", "start", daemonPlistPath)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error starting service: %w", err)
	}
	return nil
}

// restart the service
func (i *Install) restartService() error {
	err := i.stopService()
	if err != nil {
		return err
	}

	// Delay 3 seconds
	time.Sleep(3 * time.Second)

	return i.startService()
}

// getLoggedInUsers returns a map of UID to username for currently logged-in users
func getLoggedInUsers() (map[string]string, error) {
	cmd := exec.Command("who")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("could not get logged-in users: %w", err)
	}

	uidToUser := make(map[string]string)
	lines := string(output)

	// Parse who output - format is typically: username tty date time
	var users []string
	for _, line := range splitLines(lines) {
		if len(line) > 0 {
			// First field is the username
			fields := splitFields(line)
			if len(fields) > 0 {
				users = append(users, fields[0])
			}
		}
	}

	// Get UIDs for each unique user
	seen := make(map[string]bool)
	for _, user := range users {
		if seen[user] {
			continue
		}
		seen[user] = true

		cmd = exec.Command("id", "-u", user)
		output, err = cmd.Output()
		if err != nil {
			continue // Skip users we can't get UID for
		}
		uid := string(output)
		uid = trimSpace(uid)
		if uid != "" && uid != "0" { // Skip root
			uidToUser[uid] = user
		}
	}

	return uidToUser, nil
}

// bootstrapUserAgents bootstraps the LaunchAgent for all currently logged-in users
func bootstrapUserAgents() {
	uidToUser, err := getLoggedInUsers()
	if err != nil {
		fmt.Printf("Warning: could not get logged-in users: %v\n", err)
		return
	}

	if len(uidToUser) == 0 {
		fmt.Println("No users currently logged in")
		return
	}

	successCount := 0
	for uid, username := range uidToUser {
		domain := fmt.Sprintf("gui/%s", uid)
		// Run as the user, not as root, so the agent inherits user privileges
		cmd := exec.Command("sudo", "-u", username, "launchctl", "bootstrap", domain, agentPlistPath)
		err := cmd.Run()
		if err != nil {
			// Ignore errors - may already be loaded
			fmt.Printf("Note: Could not bootstrap user helper for %s (may already be running): %v\n", username, err)
			continue
		}
		successCount++
	}

	if successCount > 0 {
		fmt.Printf("Started user helper for %d logged-in user(s)\n", successCount)
	}
}

// unloadUserAgents unloads the LaunchAgent for all currently logged-in users
func unloadUserAgents() {
	uidToUser, err := getLoggedInUsers()
	if err != nil {
		// Can't get users, try to kill any user-helper processes directly
		exec.Command("pkill", "-f", "uem-agent.*--user-helper").Run()
		return
	}

	for uid, username := range uidToUser {
		domain := fmt.Sprintf("gui/%s", uid)
		target := fmt.Sprintf("%s/com.tenebris.uem-agent-user", domain)
		// Run as the user to bootout from their domain
		cmd := exec.Command("sudo", "-u", username, "launchctl", "bootout", target)
		_ = cmd.Run() // Best effort, ignore errors
	}

	// Give processes time to exit
	time.Sleep(2 * time.Second)

	// Force kill any remaining user-helper processes
	exec.Command("pkill", "-f", "uem-agent.*--user-helper").Run()
}

// Helper functions for string processing
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func splitFields(s string) []string {
	var fields []string
	var current []byte
	inField := false

	for i := 0; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\t' {
			if inField {
				fields = append(fields, string(current))
				current = nil
				inField = false
			}
		} else {
			current = append(current, s[i])
			inField = true
		}
	}
	if inField {
		fields = append(fields, string(current))
	}
	return fields
}

func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}
