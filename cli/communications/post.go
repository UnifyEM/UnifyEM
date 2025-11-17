/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package communications

import (
	"encoding/json"
	"fmt"
)

// Post sends a JSON payload to the specified endpoint and returns the response body.
func (c *Communications) Post(endpoint string, payload interface{}) (int, []byte, error) {
	var jsonData []byte
	var err error

	// Serialize the payload to JSON if it's not nil
	if payload != nil {
		jsonData, err = json.Marshal(payload)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to serialize request: %w", err)
		}
	}

	// Use the common sendRequest function to send the POST request
	return c.sendRequest("POST", endpoint, jsonData)
}
