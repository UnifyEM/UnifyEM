//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package install

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

type Install struct {
	config *global.AgentConfig
	logger interfaces.Logger
}

func New(config *global.AgentConfig, logger interfaces.Logger) *Install {
	return &Install{config: config, logger: logger}
}

// Check displays the current configuration
func (i *Install) Check() {

	// Display the configuration
	fmt.Printf("Reg Token: %s\n", i.config.AP.Get(global.ConfigRegToken).String())
	fmt.Printf("Server URL: %s\n", i.config.AP.Get(global.ConfigServerURL).String())
	fmt.Printf("Agent ID: %s\n", i.config.AP.Get(global.ConfigAgentID).String())
	fmt.Printf("\n")

	acDump, err := i.config.AC.Dump()
	if err != nil {
		fmt.Printf("Error dumping AC configuration: %v\n", err)
		return
	}
	fmt.Printf("AC: %v\n\n", acDump)

	apDump, err := i.config.AP.Dump()
	if err != nil {
		fmt.Printf("Error dumping AP configuration: %v\n", err)
		return
	}
	fmt.Printf("AP: %v\n\n", apDump)
}

func (i *Install) Install(key string) error {
	var err error

	// Check that a key is provided
	if key == "" {
		return errors.New("key is required for installation")
	}

	// Save the key
	i.config.AP.Set(global.ConfigRegToken, key)
	i.config.AP.Set(global.ConfigRefreshToken, "")
	i.config.AP.Set(global.ConfigServerURL, "")
	err = i.config.Checkpoint()
	if err != nil {
		return err
	}

	// Call the private function for os specific install
	return i.installService()
}

// Stop the service, but do not disable or uninstall it
func (i *Install) Stop() error {
	// Call the private function for os specific stop
	return i.stopService()
}

func (i *Install) Uninstall() error {
	// Call the private function for os specific uninstall
	return i.uninstallService(true)
}

func (i *Install) Upgrade() error {
	var err error

	// Call the os specific upgrade
	err = i.upgradeService()
	if err != nil {
		i.logger.Errorf(8600, "upgrade failed, attempting recovery: %s", err.Error())

		// Attempt to uninstall the service
		err = i.recoverUninstall()
		if err != nil {
			// On Windows, if the service exists and can't be removed, that's a problem
			// On Linux or macOS the binary might be locked, but it's worth trying to continue
			if runtime.GOOS == "windows" {
				if !strings.Contains(err.Error(), "service does not exist") {
					i.logger.Errorf(8601, "unable to uninstall the existing service: %s", err.Error())
					return err
				}
			}
		}

		// sleep for 5 seconds
		time.Sleep(5 * time.Second)

		// Attempt to install the service
		err = i.recoverInstall()
		if err != nil {
			i.logger.Errorf(8602, "recovery failed: %s", err.Error())
			return err
		}
		i.logger.Info(8603, "upgrade recovery was successful", nil)
	}
	return nil
}

func (i *Install) recoverUninstall() error {
	var err error

	// Try to stop the service
	err = i.stopService()
	if err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	err = i.uninstallService(false)
	if err != nil {
		return fmt.Errorf("uninstall failed: %w", err)
	}
	return nil
}

func (i *Install) recoverInstall() error {
	var err error
	var origFail string

	// attempt to install the service
	err = i.installService()
	if err != nil {
		origFail = err.Error()

		// sleep for 10 seconds
		time.Sleep(10 * time.Second)

		// try again
		err = i.installService()
		if err != nil {
			return fmt.Errorf("install failed twice: [%s] [%s]", origFail, err.Error())
		}
	}
	return nil
}

func (i *Install) ReKey(key string) error {

	// Check that a key is provided
	if key == "" {
		return errors.New("key is required")
	}

	// Save the new registration token, clear the refresh token and URL
	i.config.AP.Set(global.ConfigRegToken, key)
	i.config.AP.Set(global.ConfigRefreshToken, "")
	i.config.AP.Set(global.ConfigServerURL, "")

	err := i.config.Checkpoint()
	if err != nil {
		return err
	}

	// Attempt to restart the service, but it could already be stopped
	_ = i.restartService()
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(in *os.File) {
		_ = in.Close()
	}(in)

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	err = out.Close()
	if err != nil {
		return err
	}

	return nil
}
