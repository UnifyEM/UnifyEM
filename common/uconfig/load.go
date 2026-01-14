/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package uconfig

import (
	"encoding/json"
	"fmt"
	"os"
)

// Load configuration from specified file
func (c *UConfig) loadFile() error {
	c.Init()

	// Open the config file
	file, err := os.Open(c.file)
	if err != nil {
		return fmt.Errorf("error opening file %s: %v\n", c.file, err)
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Decode the JSON data into the struct
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&c)
	if err != nil {
		return fmt.Errorf("deserialization error: %v\n", err)
	}
	return nil
}
