//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

// Code for operating systems other than windows
//go:build linux || darwin

package ulogger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

// UEMLogger is different for eah OS because of the different loggers used
type UEMLogger struct {
	fileHandle       *os.File
	logfile          string
	logStdout        bool
	logWindowsEvents bool // Ignored on non-Windows systems
	debug            bool
	prefix           string
	retainDays       int
	currentLogDate   string
}

// New creates a new instance of UEMLogger.
func (u *UEMLogger) osNew() (*UEMLogger, error) {
	var err error
	var fh *os.File

	if u.logfile != "" {

		// Sanitize the file path
		u.logfile = filepath.Clean(u.logfile)

		// Create the directory if it doesn't exist
		dir := filepath.Dir(u.logfile)
		if err = os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Check if the log file exists
		if _, err = os.Stat(u.logfile); err == nil {
			// Get the modification date of the existing log file
			var fileInfo os.FileInfo
			fileInfo, err = os.Stat(u.logfile)
			if err != nil {
				return nil, fmt.Errorf("failed to get log file info: %w", err)
			}
			u.currentLogDate = fileInfo.ModTime().Format("20060102")
		} else {
			// Set currentLogDate to the current date if the log file does not exist
			u.currentLogDate = time.Now().Format("20060102")
		}

		fh, err = os.OpenFile(u.logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			u.fileHandle = nil
			// If unable to log to file, force stdout logging
			u.logStdout = true
		} else {
			u.fileHandle = fh

			// Attempt to set the file mode to 0644 on a best-effort basis
			_ = os.Chmod(u.logfile, 0644)
		}
	} else {
		// If no log file is specified, force stdout logging
		u.logStdout = true
	}
	return u, nil
}

// Close closes the logger.
func (u *UEMLogger) Close() {
	if u.fileHandle != nil {
		_ = u.fileHandle.Sync()
		_ = u.fileHandle.Close()
	}
}

// formatMessage formats the log message with a timestamp.
func (u *UEMLogger) formatMessage(eid uint32, level string, message string, fields interfaces.Fields) string {
	msg := fmt.Sprintf("%s %s [%s] %04d %s",
		time.Now().Format("2006-01-02 15:04:05"),
		u.prefix, level, eid, message)

	if fields != nil {
		msg += ": " + fields.ToText()
	}

	return msg
}

// writeLog writes a log message and handles rotation if necessary.
func (u *UEMLogger) writeLog(eid uint32, level string, message string, fields interfaces.Fields) {

	// Rotate logs if necessary
	err := u.rotateLogs()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "log rotation error: %s\n", err.Error())
	}

	tmp := u.formatMessage(eid, level, message, fields) + "\n"

	//  Write and flush
	if u.fileHandle != nil {
		_, _ = u.fileHandle.WriteString(tmp)
		_ = u.fileHandle.Sync()
	}

	if u.logStdout {
		_, _ = os.Stdout.Write([]byte(tmp))
	}
}

// Debug logs a debug message.
func (u *UEMLogger) Debug(eid uint32, message string, fields interfaces.Fields) {
	u.writeLog(eid, "DEBUG", message, fields)
}

// Info logs an informational message.
func (u *UEMLogger) Info(eid uint32, message string, fields interfaces.Fields) {
	u.writeLog(eid, "INFO", message, fields)
}

// Warning logs a warning message.
func (u *UEMLogger) Warning(eid uint32, message string, fields interfaces.Fields) {
	u.writeLog(eid, "WARNING", message, fields)
}

// Error logs an error message.
func (u *UEMLogger) Error(eid uint32, message string, fields interfaces.Fields) {
	u.writeLog(eid, "ERROR", message, fields)
}

// Fatal logs a fatal error message.
func (u *UEMLogger) Fatal(eid uint32, message string, fields interfaces.Fields) {
	u.writeLog(eid, "FATAL", message, fields)
}

// Debugf logs a formatted debug message.
func (u *UEMLogger) Debugf(eid uint32, format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	u.writeLog(eid, "DEBUG", message, nil)
}

// Infof logs a formatted informational message.
func (u *UEMLogger) Infof(eid uint32, format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	u.writeLog(eid, "INFO", message, nil)
}

// Warningf logs a formatted warning message.
func (u *UEMLogger) Warningf(eid uint32, format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	u.writeLog(eid, "WARNING", message, nil)
}

// Errorf logs a formatted error message.
func (u *UEMLogger) Errorf(eid uint32, format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	u.writeLog(eid, "ERROR", message, nil)
}

// Fatalf logs a formatted fatal message.
func (u *UEMLogger) Fatalf(eid uint32, format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	u.writeLog(eid, "FATAL", message, nil)
}
