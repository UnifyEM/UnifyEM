/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package uconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"strconv"
)

// Save configuration to specified file
func (c *UConfig) saveFile() error {

	// Open the file for writing (create if not exists, truncate if exists)
	file, err := os.OpenFile(c.file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Set the owner and permissions
	setOwnerPermissions(c.file)

	// Create a JSON encoder and write the struct to the file
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty print with indentation
	err = encoder.Encode(&c)
	if err != nil {
		return fmt.Errorf("could not encode to JSON: %w", err)
	}
	return nil
}

// setOwnerPermissions sets the owner and permissions for the configuration file
// on a best-effort basis.  Errors are ignored because they are not fatal.
func setOwnerPermissions(filename string) {

	// Ensure the file has the correct permissions
	_ = os.Chmod(filename, 0600)

	// Get the UID and GID for root
	rootUser, err := user.Lookup("root")
	if err != nil {
		return
	}

	uid, err := strconv.Atoi(rootUser.Uid)
	if err != nil {
		return
	}

	gid, err := strconv.Atoi(rootUser.Gid)
	if err != nil {
		return
	}

	// Set the ownership of the file to root:root
	_ = os.Chown(filename, uid, gid)
}
