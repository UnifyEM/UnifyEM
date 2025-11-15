/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package global

import (
	"encoding/json"
	"fmt"
)

func Pretty(v interface{}) {

	// Marshal the interface into a JSON string with indentation
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling to JSON: %v\n", err)
		return
	}

	// Print the pretty JSON string
	fmt.Println(string(jsonData))
}
