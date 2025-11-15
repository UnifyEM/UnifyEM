/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package communications

import "github.com/UnifyEM/UnifyEM/cli/util"

// Get sends a GET request to the specified endpoint and returns the response body.
func (c *Communications) Get(endpoint string) (int, []byte, error) {
	return c.sendRequest("GET", endpoint, nil)
}

// GetQuery accepts pairs and turns them into query parameters for a GET request to the specified endpoint
func (c *Communications) GetQuery(endpoint string, pairs *util.NVPairs) (int, []byte, error) {
	query := ""

	// Iterate through the pairs and add them to the URL as query parameters
	if pairs != nil {
		for n, v := range pairs.Pairs {
			if query == "" {
				query += "?"
			} else {
				query += "&"
			}
			query += n + "=" + v
		}
	}
	return c.sendRequest("GET", endpoint+query, nil)
}
