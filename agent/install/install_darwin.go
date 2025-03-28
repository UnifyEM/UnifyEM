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

const binaryPath = "/usr/local/bin/"
const plistPath = "/Library/LaunchDaemons/com.tenebris.uem-agent.plist"

// Note that this must also be changed if binaryPath or global.UnixBinaryName are changed
//
//goland:noinspection HttpUrlsUsage
const plistContent = `
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

// Install the service
func (i *Install) installService() error {

	// Get the path of the current executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find executable path: %w", err)
	}

	// Set the target path
	targetPath := binaryPath + global.UnixBinaryName

	// Copy the executable to the target directory
	if exePath != targetPath {
		err = copyFile(exePath, targetPath)
		if err != nil {
			return fmt.Errorf("error copying file %s to %s: %v", exePath, targetPath, err)
		}
	}
	fmt.Printf("Binary copied to %s\n", targetPath)

	// Set the proper permissions on the binary
	err = os.Chmod(binaryPath+global.UnixBinaryName, 0700)
	if err != nil {
		return fmt.Errorf("could not set permissions on binary: %w", err)
	}

	// Set the owner of the binary to root
	cmd := exec.Command("chown", "root:wheel", binaryPath+global.UnixBinaryName)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("could not set owner of binary to root: %w", err)
	}

	// Create the plist file
	err = i.createPlist()
	if err != nil {
		return err
	}

	// Load the Launch Daemon
	cmd = exec.Command("launchctl", "load", plistPath)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("could not load launch daemon: %w", err)
	}

	return nil
}

// Uninstall the service
func (i *Install) uninstallService(removeData bool) error {

	// Unload the Launch Daemon
	cmd := exec.Command("launchctl", "unload", plistPath)
	err := cmd.Run()
	if err != nil {
		// Delay 30 seconds and try again
		time.Sleep(30 * time.Second)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("could not unload launch daemon: %w", err)
		}
	}

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

	// Remove the plist file
	err = os.Remove(plistPath)
	if err != nil {
		return fmt.Errorf("could not remove plist file: %w", err)
	}

	// Remove the service binary
	err = os.Remove(binaryPath + global.UnixBinaryName)
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

// createPlist creates the Launch Daemon plist file
func (i *Install) createPlist() error {
	err := os.WriteFile(plistPath, []byte(plistContent), 0644)
	if err != nil {
		return fmt.Errorf("could not write plist file: %w", err)
	}

	fmt.Printf("Plist file created at: %s\n", plistPath)
	return nil
}

// stopService stops the service
func (i *Install) stopService() error {
	cmd := exec.Command("launchctl", "stop", plistPath)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error stopping service: %w", err)
	}
	return nil
}

// startService starts the service
func (i *Install) startService() error {
	cmd := exec.Command("launchctl", "start", plistPath)
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
