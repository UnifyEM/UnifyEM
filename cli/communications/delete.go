/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package communications

// Delete sends a DELETE request to the specified endpoint and returns the response body.
func (c *Communications) Delete(endpoint string) (int, []byte, error) {
	return c.sendRequest("DELETE", endpoint, nil)
}
