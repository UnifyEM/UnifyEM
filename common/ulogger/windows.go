/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

// Code for windows
//go:build windows

package ulogger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows/svc/eventlog"

	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

// windowsEID is the event ID for the custom event log source
// Using other event IDs will create messy log entries unless a DLL with
// messages strings is created and registered with the event log source
const windowsEID = 1

type UEMLogger struct {
	logger           *eventlog.Log
	fileHandle       *os.File
	logfile          string
	logStdout        bool
	logWindowsEvents bool
	debug            bool
	prefix           string
	retainDays       int
	currentLogDate   string
}

// New creates a new instance of UEMLogger.
func (u *UEMLogger) osNew() (*UEMLogger, error) {
	var err error
	var fh *os.File

	if u.logWindowsEvents {
		_ = eventlog.InstallAsEventCreate(u.prefix, eventlog.Info|eventlog.Warning|eventlog.Error)
	}

	u.logger, err = eventlog.Open(u.prefix)
	if err != nil {
		u.logger = nil
	}

	if u.logfile != "" {
		u.logfile = filepath.Clean(u.logfile)
		dir := filepath.Dir(u.logfile)
		if err = os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Check if the log file exists and get its modification date
		if fileInfo, err := os.Stat(u.logfile); err == nil {
			u.currentLogDate = fileInfo.ModTime().Format("20060102")
		} else {
			u.currentLogDate = time.Now().Format("20060102")
		}

		fh, err = os.OpenFile(u.logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			if u.logger != nil {
				_ = u.logger.Error(windowsEID, fmt.Sprintf("failed to open log file: %s", err.Error()))
			}
			u.fileHandle = nil
			u.logStdout = true
		} else {
			u.fileHandle = fh
		}
	}
	return u, nil
}

func (u *UEMLogger) Close() {
	if u.logger != nil {
		_ = u.logger.Close()
	}
	if u.fileHandle != nil {
		_ = u.fileHandle.Sync()
		_ = u.fileHandle.Close()
	}
}

func (u *UEMLogger) formatMessage(eid uint32, level string, message string, fields interfaces.Fields) string {
	msg := fmt.Sprintf("[%s] %04d %s", level, eid, message)
	if fields != nil {
		msg += ": " + fields.ToText()
	}
	return msg
}

func (u *UEMLogger) logMessage(eid uint32, level string, message string, fields interfaces.Fields) {

	// Rotate logs if necessary
	err := u.rotateLogs()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "log rotation error: %s\n", err.Error())
	}

	formattedMessage := u.formatMessage(eid, level, message, fields)
	if u.logger != nil {
		switch level {
		case "DEBUG", "INFO":
			_ = u.logger.Info(windowsEID, formattedMessage)
		case "WARNING":
			_ = u.logger.Warning(windowsEID, formattedMessage)
		case "ERROR", "FATAL":
			_ = u.logger.Error(windowsEID, formattedMessage)
		}
	}

	tmp := fmt.Sprintf("%s %s %s\r\n",
		time.Now().Format("2006-01-02 15:04:05"),
		u.prefix,
		formattedMessage)

	//  Write and flush
	if u.fileHandle != nil {
		_, _ = u.fileHandle.WriteString(tmp)
		_ = u.fileHandle.Sync()
	}

	if u.logStdout {
		_, _ = os.Stdout.Write([]byte(tmp))
	}
}

func (u *UEMLogger) Debug(eid uint32, message string, fields interfaces.Fields) {
	if u.debug {
		u.logMessage(eid, "DEBUG", message, fields)
	}
}

func (u *UEMLogger) Info(eid uint32, message string, fields interfaces.Fields) {
	u.logMessage(eid, "INFO", message, fields)
}

func (u *UEMLogger) Warning(eid uint32, message string, fields interfaces.Fields) {
	u.logMessage(eid, "WARNING", message, fields)
}

func (u *UEMLogger) Error(eid uint32, message string, fields interfaces.Fields) {
	u.logMessage(eid, "ERROR", message, fields)
}

func (u *UEMLogger) Fatal(eid uint32, message string, fields interfaces.Fields) {
	u.logMessage(eid, "FATAL", message, fields)
}

func (u *UEMLogger) Debugf(eid uint32, format string, v ...any) {
	if u.debug {
		message := fmt.Sprintf(format, v...)
		u.logMessage(eid, "DEBUG", message, nil)
	}
}

func (u *UEMLogger) Infof(eid uint32, format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	u.logMessage(eid, "INFO", message, nil)
}

func (u *UEMLogger) Warningf(eid uint32, format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	u.logMessage(eid, "WARNING", message, nil)
}

func (u *UEMLogger) Errorf(eid uint32, format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	u.logMessage(eid, "ERROR", message, nil)
}

func (u *UEMLogger) Fatalf(eid uint32, format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	u.logMessage(eid, "FATAL", message, nil)
}
