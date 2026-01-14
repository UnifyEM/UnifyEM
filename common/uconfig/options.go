/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package uconfig

import (
	"fmt"
	"os"
)

//goland:noinspection GoUnusedExportedFunction
func WithLoad(filename string) func(*UConfig) error {
	return func(c *UConfig) error {
		return c.Load(filename)
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithLoadOrCreate(filename string) func(*UConfig) error {
	return func(c *UConfig) error {
		err := c.Load(filename)
		if err != nil {
			return c.Save(filename)
		}
		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithFind(filenames []string) func(*UConfig) error {
	return func(c *UConfig) error {
		// Iterate through the possible configuration files
		for _, filename := range filenames {
			if _, err := os.Stat(filename); err == nil {
				return c.Load(filename)
			}
		}
		return fmt.Errorf("no configuration file found")
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithFindOrCreate(filenames []string) func(*UConfig) error {
	return func(c *UConfig) error {
		// Iterate through the possible configuration files
		for _, filename := range filenames {
			if _, err := os.Stat(filename); err == nil {
				return c.Load(filename)
			}
		}

		// file was not found, so create it
		// Iterate through the possible configuration files
		for _, filename := range filenames {
			// Attempt to write
			file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
			if err == nil {
				// Success
				_ = file.Close()
				return c.Save(filename)
			}
		}
		return fmt.Errorf("could not create configuration file")
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithWindowsRegistry(key string) func(*UConfig) error {
	return func(c *UConfig) error {
		c.windowsRegistry = true
		c.windowsRegistryKey = key
		return c.Load("")
	}
}
