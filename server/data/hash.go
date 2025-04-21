//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package data

import (
	"os"
	"strings"

	"github.com/UnifyEM/UnifyEM/server/global"
)

// getHashOfFile builds a path to the file in our file directory, checks if it exists, and if so
// returns the base64 encoded SHA256 hash of the file. Otherwise, it returns an empty string
func (d *Data) getHashOfFile(file string) string {
	if file == "" {
		return ""
	}

	// If there are any path separators of any kind in the filename, reject the file
	if strings.ContainsAny(file, string(os.PathSeparator)+"/"+"\\") {
		return ""
	}

	// If the file starts with a ., reject it
	if strings.HasPrefix(file, ".") {
		return ""
	}

	// Get the path to our file server directory
	path := d.conf.SC.Get(global.ConfigFilesPath).String()

	// Ensure that it ends with a path separator
	if !strings.HasSuffix(path, string(os.PathSeparator)) {
		path += string(os.PathSeparator)
	}

	// Append the filename to the path
	path = path + file

	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ""
	}

	// Get the hash of the file
	return d.hasher.SHA256File(path).Base64()
}
