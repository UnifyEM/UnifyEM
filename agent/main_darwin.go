/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

//go:build darwin

package main

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"time"

	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/agent/userdata"
	"github.com/UnifyEM/UnifyEM/agent/userhelper"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
	"github.com/UnifyEM/UnifyEM/common/ulogger"
)

var userDataListener *userdata.UserDataListener

// checkUserHelperMode checks if --user-helper flag is present and runs user helper mode if so.
// Returns true if user helper mode was activated (and will exit), false to continue normal execution.
func checkUserHelperMode() bool {
	if len(os.Args) > 1 && os.Args[1] == global.UserHelperFlag {
		// Parse collection interval from args if provided
		interval := global.DefaultCollectionInterval
		if len(os.Args) > 3 && os.Args[2] == global.CollectionIntervalFlag {
			if val, err := strconv.Atoi(os.Args[3]); err == nil && val > 0 {
				interval = val
			}
		}
		runUserHelper(interval)
		os.Exit(0)
		return true // Never reached, but for clarity
	}
	return false
}

// initUserDataListener starts the user data listener for macOS
func initUserDataListener(log interfaces.Logger) {
	userDataListener = userdata.New(log)
	if err := userDataListener.Start(); err != nil {
		log.Errorf(8003, "Failed to start user data listener: %v", err)
		// Continue without it - will fall back to existing methods
	} else {
		// Clean stale data periodically
		go func() {
			ticker := time.NewTicker(15 * time.Minute)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					userDataListener.CleanStaleData(30 * time.Minute)
				}
			}
		}()
	}
}

// cleanupUserDataListener stops the user data listener
func cleanupUserDataListener(log interfaces.Logger) {
	if userDataListener != nil {
		if err := userDataListener.Stop(); err != nil {
			log.Errorf(8007, "Error stopping user data listener: %v", err)
		}
	}
}

// getUserDataSource returns the user data listener for use in command functions
func getUserDataSource() *userdata.UserDataListener {
	return userDataListener
}

// runUserHelper is called when --user-helper flag is detected
func runUserHelper(collectionInterval int) {
	// Minimal setup - no privilege escalation needed
	username := getCurrentUsername()

	// Create per-user log file
	logPath := fmt.Sprintf("/tmp/uem-agent-user-%s.log", username)

	// Create logger (log to /tmp since we don't have /var/log access)
	logger, err := ulogger.New(
		ulogger.WithPrefix("uem-agent-user"),
		ulogger.WithLogFile(logPath),
		ulogger.WithLogStdout(false),
		ulogger.WithRetention(7),
		ulogger.WithDebug(global.Debug))

	if err != nil {
		fmt.Printf("Error creating logger: %v\n", err)
		os.Exit(1)
	}

	logger.Infof(3200, "Starting user-helper mode for user: %s (interval: %d seconds)",
		username, collectionInterval)

	// Load minimal config (or use defaults)
	config := &global.AgentConfig{
		// Minimal config needed for status collection
	}

	// Create and run user helper
	helper := userhelper.New(logger, config, collectionInterval)

	if err := helper.Run(); err != nil {
		logger.Errorf(3201, "User-helper error: %v", err)
		os.Exit(1)
	}
}

// getCurrentUsername returns the current user's username
func getCurrentUsername() string {
	u, err := user.Current()
	if err != nil {
		return "unknown"
	}
	return u.Username
}
