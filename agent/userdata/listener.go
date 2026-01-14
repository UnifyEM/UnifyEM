/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

// Package userdata manages user-context data reception from user-helper processes
package userdata

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/UnifyEM/UnifyEM/agent/functions/status"
	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

// UserDataListener manages the Unix socket server for receiving user-context data
//
//goland:noinspection GoNameStartsWithPackageName
type UserDataListener struct {
	logger          interfaces.Logger
	listener        net.Listener
	mu              sync.RWMutex
	consoleUserData status.UserContextData // Only store console user data
	hasData         bool
	running         bool
}

// New creates a new UserDataListener instance
func New(logger interfaces.Logger) *UserDataListener {
	return &UserDataListener{
		logger: logger,
	}
}

// Start begins listening on the Unix socket
func (l *UserDataListener) Start() error {
	// Remove stale socket if it exists
	_ = os.Remove(global.SocketPath)

	listener, err := net.Listen("unix", global.SocketPath)
	if err != nil {
		return fmt.Errorf("failed to create socket listener: %w", err)
	}

	// Set permissions so user processes can connect
	if err := os.Chmod(global.SocketPath, global.SocketPerms); err != nil {
		_ = listener.Close()
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	l.listener = listener
	l.running = true
	l.logger.Infof(3100, "User data listener started on %s", global.SocketPath)

	// Start accepting connections in background
	go l.acceptLoop()

	return nil
}

// acceptLoop handles incoming connections
func (l *UserDataListener) acceptLoop() {
	for l.running {
		conn, err := l.listener.Accept()
		if err != nil {
			if l.running {
				l.logger.Errorf(3101, "Error accepting connection: %v", err)
			}
			continue
		}

		// Handle each connection in a goroutine
		go l.handleConnection(conn)
	}
}

// handleConnection processes a single user-helper connection
func (l *UserDataListener) handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	// Set read deadline
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	var data status.UserContextData
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&data); err != nil {
		l.logger.Errorf(3102, "Error decoding user data: %v", err)
		return
	}

	// Store the received data (console user only)
	l.mu.Lock()
	l.consoleUserData = data
	l.hasData = true
	l.mu.Unlock()

	l.logger.Debugf(3103, "Received console user data from %s: screen_lock=%s, delay=%s",
		data.Username, data.ScreenLock, data.ScreenLockDelay)
}

// GetConsoleUserData retrieves stored user-context data for the console user
func (l *UserDataListener) GetConsoleUserData() (status.UserContextData, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.consoleUserData, l.hasData
}

// CleanStaleData removes user data older than the specified duration
func (l *UserDataListener) CleanStaleData(maxAge time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.hasData {
		return
	}

	cutoff := time.Now().Add(-maxAge)
	if l.consoleUserData.Timestamp.Before(cutoff) {
		l.logger.Debugf(3104, "Removed stale user data for %s", l.consoleUserData.Username)
		l.hasData = false
		l.consoleUserData = status.UserContextData{}
	}
}

// Stop closes the listener and cleans up
func (l *UserDataListener) Stop() error {
	l.running = false

	if l.listener != nil {
		if err := l.listener.Close(); err != nil {
			return err
		}
	}

	// Clean up socket file
	_ = os.Remove(global.SocketPath)

	l.logger.Infof(3105, "User data listener stopped")
	return nil
}
