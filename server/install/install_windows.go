//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

// Windows specific functions
//go:build windows

package install

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

	"github.com/UnifyEM/UnifyEM/common/uemservice/privcheck"
	"github.com/UnifyEM/UnifyEM/server/global"
)

// SERVICE_FAILURE_ACTIONS structure
//
//goland:noinspection GoSnakeCaseUsage
type SERVICE_FAILURE_ACTIONS struct {
	DwResetPeriod uint32
	LpRebootMsg   *uint16
	LpCommand     *uint16
	CActions      uint32
	LpActions     *SC_ACTION
}

// SC_ACTION_TYPE constants
//
//goland:noinspection ALL
const (
	SC_ACTION_NONE        = 0
	SC_ACTION_RESTART     = 1
	SC_ACTION_REBOOT      = 2
	SC_ACTION_RUN_COMMAND = 3
)

// SC_ACTION structure
//
//goland:noinspection GoSnakeCaseUsage
type SC_ACTION struct {
	Type  uint32
	Delay uint32
}

//goland:noinspection GoSnakeCaseUsage
const SERVICE_CONFIG_FAILURE_ACTIONS = 2

// Install the service
func (i *Install) installService() error {

	// Get the executable path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %w", err)
	}

	// Define the target directory and file
	targetDir := filepath.Join(os.Getenv("ProgramFiles"), global.Name)
	targetPath := filepath.Join(targetDir, global.WindowsBinaryName)

	// Create the directory if it doesn't exist
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		err := os.MkdirAll(targetDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
	}

	// Copy the executable to the target directory
	if exePath != targetPath {
		err = copyFile(exePath, targetPath)
		if err != nil {
			return fmt.Errorf("error copying file %s to %s: %v", exePath, targetPath, err)
		}
	}
	fmt.Printf("Binary copied to %s\n", targetPath)

	// Install the service
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("error connecting to service manager: %w", err)
	}
	defer func(m *mgr.Mgr) {
		_ = m.Disconnect()
	}(m)

	service, err := m.OpenService(global.Name)
	if err == nil {
		_ = service.Close()
		return fmt.Errorf("service %s already exists", global.Name)
	}

	service, err = m.CreateService(global.Name, targetPath, mgr.Config{
		DisplayName: global.Description,
		Description: global.Description,
		StartType:   mgr.StartAutomatic,
		ServiceType: windows.SERVICE_WIN32_OWN_PROCESS,
	}, "is", "auto-started")
	if err != nil {
		return fmt.Errorf("error creating service: %w", err)
	}
	defer func(service *mgr.Service) {
		_ = service.Close()
	}(service)

	fmt.Println("Windows service created")

	// Set the service failure actions
	err = setServiceFailureActions(service.Handle)
	if err != nil {
		return fmt.Errorf("could not set failure actions: %w", err)
	}
	fmt.Println("Windows service failure actions set")

	// Start the service
	err = service.Start()
	if err != nil {
		return fmt.Errorf("error starting service: %w", err)
	}

	fmt.Println("Windows service started")

	return nil
}

// setServiceFailureActions sets the failure actions for the service.
func setServiceFailureActions(serviceHandle windows.Handle) error {
	actions := []SC_ACTION{
		{Type: SC_ACTION_RESTART, Delay: 10000}, // Restart the service after 10 seconds
		{Type: SC_ACTION_RESTART, Delay: 10000}, // Restart the service after 10 seconds
		{Type: SC_ACTION_RESTART, Delay: 60000}, // Restart the service after 60 seconds
	}

	// Prepare the SERVICE_FAILURE_ACTIONS structure
	failureActions := SERVICE_FAILURE_ACTIONS{
		DwResetPeriod: 86400, // Reset the failure count after 1 day (86400 seconds)
		LpCommand:     nil,
		LpRebootMsg:   nil,
		CActions:      uint32(len(actions)),
		LpActions:     &actions[0],
	}

	// Call the ChangeServiceConfig2 function to set the failure actions
	r1, _, e1 := syscall.NewLazyDLL("advapi32.dll").NewProc("ChangeServiceConfig2W").Call(
		uintptr(serviceHandle),
		uintptr(SERVICE_CONFIG_FAILURE_ACTIONS),
		uintptr(unsafe.Pointer(&failureActions)),
	)

	if r1 == 0 {
		return fmt.Errorf("ChangeServiceConfig2W failed: %v", e1)
	}

	return nil
}

// Uninstall the service
func (i *Install) uninstallService(removeData bool) error {

	// Connect to the service manager
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("error connecting to service manager: %w", err)
	}
	defer func(m *mgr.Mgr) {
		_ = m.Disconnect()
	}(m)

	// Open the service
	service, err := m.OpenService(global.Name)
	if err != nil {
		return fmt.Errorf("error opening service: %w", err)
	}
	defer func(service *mgr.Service) {
		_ = service.Close()
	}(service)

	// Stop the service if it is running
	status, err := service.Query()
	if err == nil && status.State == svc.Running {
		_, err = service.Control(svc.Stop)
		if err != nil {
			return fmt.Errorf("error stopping service: %w", err)
		}
		fmt.Println("Service stopped")
	}

	// Delete the service
	err = service.Delete()
	if err != nil {
		return fmt.Errorf("error deleting service: %w", err)
	}
	fmt.Println("Service uninstalled")

	// Define the target directory and file
	targetDir := filepath.Join(os.Getenv("ProgramFiles"), global.Name)
	targetPath := filepath.Join(targetDir, global.WindowsBinaryName)

	// Delete the binary file from Program Files
	fmt.Printf("Attempting to delete %s\n", targetPath)
	err = i.deleteFile(targetPath)
	if err != nil {
		fmt.Printf("Unable to delete binary: %v\n", err)
	} else {
		fmt.Println("Binary deleted successfully")
	}

	if removeData {
		// TODO delete the data
	}

	return nil
}

// Try to delete a file several times. The first few often fail if the service has just
// been removed
func (i *Install) deleteFile(path string) error {
	var err error

	for i := 0; i < 20; i++ {
		err := os.Remove(path)
		if err == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	// return the error from the last attempt
	return err
}

// Upgrade the service
func (i *Install) upgradeService() error {
	var err error

	fmt.Println("Uninstalling existing server...")

	// Remove the existing executable
	err = i.uninstallService(false)
	if err != nil {
		return fmt.Errorf("could not remove existing service: %w", err)
	}

	// Delay for two seconds to allow the system to release the file
	time.Sleep(1 * time.Second)

	fmt.Println("\nInstalling new server...")

	// Install the new service
	return i.installService()
}

// CheckAdmin checks if the current process is running with administrator privileges,
// and if not, it attempts to restart the process with administrator privileges.
func CheckAdmin() error {
	admin, err := privcheck.Check()
	if err != nil {
		return err
	}

	if !admin {
		fmt.Println("Not running as admin, attempting to restart with admin privileges...")
		err := runAsAdmin()
		if err != nil {
			fmt.Printf("Failed to restart as admin: %v\n", err)
			return fmt.Errorf("failed to restart as admin: %w", err)
		} else {
			os.Exit(0)
		}
	}

	return nil
}

// runAsAdmin restarts the current process with administrator privileges.
func runAsAdmin() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exePath, err := windows.UTF16PtrFromString(exe)
	if err != nil {
		return err
	}

	// Get the command-line arguments and join them into a single string.
	args := strings.Join(os.Args[1:], " ")
	argsPtr, err := windows.UTF16PtrFromString(args)
	if err != nil {
		return err
	}

	verb, err := windows.UTF16PtrFromString("runas")
	if err != nil {
		return err
	}

	// Use ShellExecute to restart the program with elevated privileges, passing the command-line arguments.
	err = windows.ShellExecute(0, verb, exePath, argsPtr, nil, windows.SW_NORMAL)
	if err != nil && err.Error() != "The operation completed successfully." {
		return err
	}

	return nil
}

// CheckRootPrivileges is an alias for CheckAdmin for compatibility
func CheckRootPrivileges() error {
	return CheckAdmin()
}

// stopService stops the service // TODO
func (i *Install) stopService() error {
	return nil
}

// startService starts the service // TODO
func (i *Install) startService() error {
	return nil
}
