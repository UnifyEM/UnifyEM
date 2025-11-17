//go:build darwin

/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

// Package userhelper provides user-context data collection for macOS
package userhelper

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"os/user"
	"strings"
	"time"

	"github.com/UnifyEM/UnifyEM/agent/functions/status"
	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

// UserHelper manages user-context data collection and transmission
type UserHelper struct {
	logger               interfaces.Logger
	config               *global.AgentConfig
	collectionInterval   time.Duration
	tccNotificationShown bool // Track if we've shown TCC notification in this run
}

// UserContextData represents user-specific context information
type UserContextData struct {
	Username        string            `json:"username"`
	Timestamp       time.Time         `json:"timestamp"`
	ScreenLock      string            `json:"screen_lock"`
	ScreenLockDelay string            `json:"screen_lock_delay"`
	RawData         map[string]string `json:"raw_data"`
}

// New creates a new UserHelper instance
func New(logger interfaces.Logger, config *global.AgentConfig, intervalSeconds int) *UserHelper {
	return &UserHelper{
		logger:             logger,
		config:             config,
		collectionInterval: time.Duration(intervalSeconds) * time.Second,
	}
}

// Run is the main loop for user-helper mode
func (h *UserHelper) Run() error {
	username := getCurrentUsername()
	h.logger.Infof(3000, "Starting user-helper mode for user %s with collection interval %v",
		username, h.collectionInterval)

	ticker := time.NewTicker(h.collectionInterval)
	defer ticker.Stop()

	// Send initial data immediately
	if err := h.collectAndSend(); err != nil {
		h.logger.Errorf(3001, "Error collecting initial data: %v", err)
	}

	// Periodic collection
	for range ticker.C {
		if err := h.collectAndSend(); err != nil {
			h.logger.Errorf(3002, "Error collecting periodic data: %v", err)
			// Continue running despite errors
		}
	}

	return nil
}

// collectAndSend gathers user-context data and sends to daemon
func (h *UserHelper) collectAndSend() error {
	// Only collect and send if this is the console user
	if !h.isConsoleUser() {
		h.logger.Debugf(3010, "Not console user, skipping data collection")
		return nil
	}

	data := h.collectUserData()
	return h.sendToDaemon(data)
}

// isConsoleUser checks if the current user is the active console user
func (h *UserHelper) isConsoleUser() bool {
	cmd := exec.Command("/usr/bin/stat", "-f", "%Su", "/dev/console")
	output, err := cmd.Output()
	if err != nil {
		h.logger.Errorf(3011, "Error checking console user: %v", err)
		return false
	}

	consoleUser := strings.TrimSpace(string(output))
	currentUser := getCurrentUsername()

	isConsole := consoleUser == currentUser
	h.logger.Debugf(3012, "Console user: %s, Current user: %s, Is console: %v",
		consoleUser, currentUser, isConsole)

	return isConsole
}

// collectUserData gathers user-specific information
func (h *UserHelper) collectUserData() status.UserContextData {
	// Create status handler to use existing collection functions
	statusHandler := status.New(h.config, h.logger, nil, nil)

	data := status.UserContextData{
		Username:  getCurrentUsername(),
		Timestamp: time.Now(),
		RawData:   make(map[string]string),
	}

	// These calls will now work because we're running in user context
	screenLockValue, screenLockErr := statusHandler.ScreenLock()
	data.ScreenLock = screenLockValue
	data.ScreenLockDelay = statusHandler.ScreenLockDelay()

	// Detect TCC permission denial
	// If ScreenLock() returned an error and the value is "unknown", it's likely a TCC issue
	if screenLockErr != nil && screenLockValue == "unknown" && !h.tccNotificationShown {
		h.handleTCCDenial(screenLockErr)
		h.tccNotificationShown = true
	}

	// Collect additional user-context data
	data.RawData["last_user"] = statusHandler.LastUser()

	return data
}

// handleTCCDenial detects TCC permission denial and shows a blocking dialog to the user
func (h *UserHelper) handleTCCDenial(err error) {
	// Check if this is specifically a TCC error (-1743)
	errMsg := err.Error()
	isTCCError := strings.Contains(errMsg, "-1743") ||
		strings.Contains(errMsg, "Not authorized to send Apple events")

	if !isTCCError {
		// Not a TCC error, just log and return
		h.logger.Warningf(3020, "Screen lock detection failed (non-TCC): %v", err)
		return
	}

	// Log the TCC denial
	h.logger.Warningf(3021, "TCC permission denied for System Events - showing user notification")

	// Show blocking dialog to user
	dialogScript := `display dialog "UEM Agent needs your permission to monitor security settings.

To enable full security monitoring:

1. Go to System Settings
2. Navigate to Privacy & Security > Automation
3. Find 'uem-agent' in the list
4. Enable 'System Events'

Without this permission, some security settings cannot be monitored." buttons {"OK"} default button "OK" with title "UEM Agent - Permission Required" with icon caution`

	cmd := exec.Command("/usr/bin/osascript", "-e", dialogScript)
	err = cmd.Run()
	if err != nil {
		h.logger.Errorf(3022, "Failed to show TCC notification dialog: %v", err)
	} else {
		h.logger.Infof(3023, "TCC notification dialog shown to user")
	}
}

// sendToDaemon sends data to daemon via Unix socket
func (h *UserHelper) sendToDaemon(data status.UserContextData) error {
	conn, err := net.DialTimeout("unix", global.SocketPath, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to daemon socket: %w", err)
	}
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	// Set write deadline
	_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	// Send JSON payload
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}

	h.logger.Debugf(3003, "Sent user data to daemon: screen_lock=%s, delay=%s",
		data.ScreenLock, data.ScreenLockDelay)

	return nil
}

// getCurrentUsername returns the current user's username
func getCurrentUsername() string {
	u, err := user.Current()
	if err != nil {
		return "unknown"
	}
	return u.Username
}
