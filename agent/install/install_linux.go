//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

// MacOS (Darin) specific functions
//go:build linux

package install

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

const (
	serviceName = "uem-agent"
	binaryPath  = "/usr/local/bin/"
	servicePath = "/etc/systemd/system"
	serviceFile = "uem-agent.service"
)

// This must also be changed if binaryPath or serviceName are changed
const serviceContent = `
[Unit]
Description=uem-agent
After=network.target
StartLimitIntervalSec=0

[Service]
WorkingDirectory=/tmp
User=root
Group=root
Restart=always
RestartSec=1
ExecStart=/usr/local/bin/uem-agent

[Install]
WantedBy=multi-user.target
Alias=uem-agent.service
`

// Install the service
func (i *Install) installService() error {

	// Get the path of the current executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find executable path: %w", err)
	}

	// Return error if servicePath doesn't exist
	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		return fmt.Errorf("%s does not exist - aborting install", servicePath)
	}

	// Set the target path
	targetPath := binaryPath + string(os.PathSeparator) + serviceName

	// Copy the executable to the target directory
	if exePath != targetPath {
		err = copyFile(exePath, targetPath)
		if err != nil {
			return fmt.Errorf("error copying file %s to %s: %w", exePath, targetPath, err)
		}
	}
	fmt.Printf("Binary copied to %s\n", targetPath)

	// Set the proper permissions on the binary
	err = os.Chmod(targetPath, 0700)
	if err != nil {
		return fmt.Errorf("could not set permissions on binary: %w", err)
	}

	// Create the service file
	err = i.createService()
	if err != nil {
		return err
	}

	// Start the service
	err = i.startService()
	if err != nil {
		return fmt.Errorf("could not start service: %w", err)
	}

	return nil
}

// Uninstall the service
func (i *Install) uninstallService(removeData bool) error {

	// Stop the service
	err := i.stopService()
	if err != nil {
		return fmt.Errorf("could not stop service: %w", err)
	}

	// Remove the service file
	err = os.Remove(servicePath + string(os.PathSeparator) + serviceFile)
	if err != nil {
		return fmt.Errorf("could not remove service file: %w", err)
	}

	// Remove the binary
	err = os.Remove(binaryPath + string(os.PathSeparator) + serviceName)
	if err != nil {
		return fmt.Errorf("could not remove binary file: %w", err)
	}

	if removeData {
		// TODO delete the data
	}

	return nil
}

// Upgrade the service
func (i *Install) upgradeService() error {

	fmt.Println("Uninstalling existing server...")

	// Remove the existing executable
	err := i.uninstallService(false)
	if err != nil {
		return fmt.Errorf("could not remove existing service: %w", err)
	}

	// Delay for two seconds to allow the system to release the file
	time.Sleep(2 * time.Second)

	fmt.Println("\nInstalling new server...")

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

// createService creates the Linux service file
func (i *Install) createService() error {
	target := servicePath + string(os.PathSeparator) + serviceFile
	err := os.WriteFile(target, []byte(serviceContent), 0644)
	if err != nil {
		return fmt.Errorf("could not write service file: %w", err)
	}

	cmd := exec.Command("systemctl", "enable", serviceName)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error enabling service: %w", err)
	}

	fmt.Printf("Service file created at: %s\n", target)
	return nil
}

// stopService stops the service
func (i *Install) stopService() error {
	cmd := exec.Command("systemctl", "stop", serviceName)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error stopping service: %w", err)
	}
	return nil
}

// startService starts the service
func (i *Install) startService() error {
	cmd := exec.Command("systemctl", "start", serviceName)
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
