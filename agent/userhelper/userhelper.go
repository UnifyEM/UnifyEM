//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

// Package userhelper provides user-context data collection for macOS
//go:build darwin

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
	logger             interfaces.Logger
	config             *global.AgentConfig
	collectionInterval time.Duration
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
	data.ScreenLock, _ = statusHandler.ScreenLock()
	data.ScreenLockDelay = statusHandler.ScreenLockDelay()

	// Collect additional user-context data
	data.RawData["last_user"] = statusHandler.LastUser()

	return data
}

// sendToDaemon sends data to daemon via Unix socket
func (h *UserHelper) sendToDaemon(data status.UserContextData) error {
	conn, err := net.DialTimeout("unix", global.SocketPath, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to daemon socket: %w", err)
	}
	defer conn.Close()

	// Set write deadline
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

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
