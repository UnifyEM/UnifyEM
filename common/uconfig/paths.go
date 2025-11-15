/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package uconfig

import (
	"os"
	"path/filepath"
)

// CreateDir attempts to create the specified directory and
// returns a bool to indicate success or failure. If the directory
// already exists that is considered a success.
func CreateDir(path string) bool {
	// Use os.MkdirAll to create the directory. It creates parents if necessary.
	err := os.MkdirAll(path, 0700)
	if err != nil {
		return false
	}
	return true
}

// CreateSubDir attempt joins the path and directory using the correct separator for the OS
// and then calls CreateDir to create the directory.
//
//goland:noinspection GoUnusedExportedFunction
func CreateSubDir(dir string, subDir string) string {
	newDir := filepath.Join(dir, subDir)
	if CreateDir(newDir) {
		return newDir
	}
	return ""
}
