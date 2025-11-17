/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package ulogger

import (
	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

// This package implements interfaces.Logger
var _ interfaces.Logger = (*UEMLogger)(nil)

// Option is a function that configures a UEMLogger
type Option func(*UEMLogger) error

// New creates a new instance of UEMLogger with the provided options
func New(options ...Option) (interfaces.Logger, error) {
	u := &UEMLogger{retainDays: 30}

	for _, option := range options {
		if err := option(u); err != nil {
			return nil, err
		}
	}

	// Call the OS-specific constructor
	return u.osNew()
}

// WithPrefix sets a process name or similar short identifier
func WithPrefix(prefix string) Option {
	return func(u *UEMLogger) error {
		u.prefix = prefix
		return nil
	}
}

// WithLogFile sets the log file for the UEMLogger
func WithLogFile(logfile string) Option {
	return func(u *UEMLogger) error {
		u.logfile = logfile
		return nil
	}
}

// WithLogStdout enables or disables logging to stdout
func WithLogStdout(logStdout bool) Option {
	return func(u *UEMLogger) error {
		u.logStdout = logStdout
		return nil
	}
}

// WithWindowsEvents enables or disables logging to the windows event log
func WithWindowsEvents(logWindowsEvents bool) Option {
	return func(u *UEMLogger) error {
		u.logWindowsEvents = logWindowsEvents
		return nil
	}
}

// WithDebug enables or disables debug logging
func WithDebug(debug bool) Option {
	return func(u *UEMLogger) error {
		u.debug = debug
		return nil
	}
}

// WithRetention sets the number of days to retain logs
func WithRetention(retainDays int) Option {
	return func(u *UEMLogger) error {
		u.retainDays = retainDays
		return nil
	}
}
