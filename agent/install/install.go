/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

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
	config    *global.AgentConfig
	logger    interfaces.Logger
	token     string
	user      string
	pass      string
	isUpgrade bool
}

// Option is a functional option for configuring Install
type Option func(*Install)

// WithConfig sets the agent configuration
func WithConfig(config *global.AgentConfig) Option {
	return func(i *Install) {
		i.config = config
	}
}

// WithLogger sets the logger
func WithLogger(logger interfaces.Logger) Option {
	return func(i *Install) {
		i.logger = logger
	}
}

// WithCredentials sets the user credentials (optional)
func WithCredentials(user, pass string) Option {
	return func(i *Install) {
		i.user = user
		i.pass = pass
	}
}

// WithToken sets the installation token (optional)
func WithToken(token string) Option {
	return func(i *Install) {
		i.token = token
	}
}

// WithUpgrade sets the upgrade flag (optional)
//
//goland:noinspection GoUnusedExportedFunction
func WithUpgrade() Option {
	return func(i *Install) {
		i.isUpgrade = true
	}
}

// New creates a new Install instance with the provided options
func New(opts ...Option) (*Install, error) {
	i := &Install{}
	for _, opt := range opts {
		opt(i)
	}

	// Validate required fields
	if i.config == nil {
		return &Install{}, errors.New("config is required")
	}
	if i.logger == nil {
		return &Install{}, errors.New("logger is required")
	}

	return i, nil
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

func (i *Install) Install() error {
	var err error

	// Check that a key is provided
	if i.token == "" {
		return errors.New("installation token is required")
	}

	// Save the key
	i.config.AP.Set(global.ConfigRegToken, i.token)
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

	// Set upgrade flag to skip service account operations
	i.isUpgrade = true

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

func (i *Install) ReKey() error {

	// Check that a key is provided
	if i.token == "" {
		return errors.New("registration token is required")
	}

	// Save the new registration token, clear the refresh token and URL
	i.config.AP.Set(global.ConfigRegToken, i.token)
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
