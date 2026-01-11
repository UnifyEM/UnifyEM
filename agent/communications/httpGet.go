/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package communications

import (
	"fmt"
	"io"
	"net/http"
)

// get sends a GET agent to the specified URL and returns the response body
// unmarshalled into schema.ServerResponse or an error.
func (c *Communications) get(server string, path string, auth bool) ([]byte, error) {

	// Build the URL with some validation
	url, err := buildURL(server, path)
	if err != nil {
		return nil, err
	}

	// Create a new HTTP GET agent
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// If authentication is required, obtain and set the bearer token
	// GetToken() will attempt refresh or registration if required
	if auth {
		token, tErr := c.GetToken()
		if tErr != nil {
			return nil, tErr
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	// Create a custom HTTP client to support CA pinning
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: c.TLSConfig(),
		},
	}

	// Perform the HTTP GET
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// If authentication is required, check for 401 status code
	if auth && resp.StatusCode == http.StatusUnauthorized {
		// Clear the token to trigger a login on the next request
		c.ClearToken()
	}

	// Check for non-200 status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s failed with status %d", url, resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
