/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package uconfig

import (
	"encoding/json"
	"fmt"
)

// Dump the configuration to the console
func (c *UConfig) Dump() {
	fmt.Printf("Current configuration:\n")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		fmt.Printf("Serialization error: %s\n", err.Error())
		return
	}
	fmt.Println(string(data))
}
