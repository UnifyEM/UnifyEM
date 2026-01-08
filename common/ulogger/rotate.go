/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package ulogger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// rotateLogs handles log rotation and deletion of old logs
func (u *UEMLogger) rotateLogs() error {
	var err error

	if u.logfile == "" {
		return nil
	}

	// Get the current date
	currentDate := time.Now().Format("20060102")

	// Check if the log file needs to be rotated
	if u.currentLogDate != currentDate {

		// Capture the date of the current log file before rotation
		previousLogDate := u.currentLogDate

		// Close the current log file
		if u.fileHandle != nil {
			_ = u.fileHandle.Sync()
			_ = u.fileHandle.Close()
		}

		// Rename the current log file
		newLogFileName := fmt.Sprintf("%s-%s", u.logfile, previousLogDate)
		err = os.Rename(u.logfile, newLogFileName)
		if err != nil {
			return fmt.Errorf("failed to rotate log file: %w", err)
		}

		// Open a new log file
		var fh *os.File
		fh, err = os.OpenFile(u.logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			u.fileHandle = nil
			u.logStdout = true
			return fmt.Errorf("failed to open new log file after rotating: %w", err)
		}
		u.fileHandle = fh

		// Attempt to set the file mode to 0644 on a best-effort basis
		_ = os.Chmod(u.logfile, 0644)

		u.currentLogDate = currentDate

		// Delete old log files
		err = u.deleteOldLogs()
		if err != nil {
			return fmt.Errorf("failed to delete old log files: %w", err)
		}
	}
	return nil
}

// deleteOldLogs deletes log files older than retainDays
func (u *UEMLogger) deleteOldLogs() error {
	if u.retainDays <= 1 {
		return nil
	}

	// Calculate the cutoff date
	cutoffDate := time.Now().AddDate(0, 0, -u.retainDays).Format("20060102")

	// Get the directory of the log files
	logDir := filepath.Dir(u.logfile)

	// List all files in the log directory
	files, err := os.ReadDir(logDir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %w", err)
	}

	fileBaseName := filepath.Base(u.logfile)

	// Delete files older than the cutoff date
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileName := file.Name()
		if len(fileName) > len(fileBaseName)+7 && fileName[:len(fileBaseName)] == fileBaseName {
			fileDate := fileName[len(fileBaseName)+1 : len(fileBaseName)+9]
			if fileDate < cutoffDate {
				err = os.Remove(filepath.Join(logDir, fileName))
				if err != nil {
					return fmt.Errorf("failed to delete old log file: %w", err)
				}
			}
		}
	}

	return nil
}
