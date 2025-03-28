//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package communications

import (
	"bytes"
	"encoding/json"
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

	// Perform the HTTP GET agent
	client := &http.Client{}
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

// post sends a POST agent with a JSON payload to the specified URL
// and returns the response body unmarshalled into schema.ServerResponse or an error.
func (c *Communications) post(server string, path string, auth bool, data any) ([]byte, error) {

	// Build the URL with some validation
	url, err := buildURL(server, path)
	if err != nil {
		return nil, err
	}

	// Marshal the struct to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Create a new HTTP POST agent
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
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

	// Perform the HTTP POST agent
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// If authentication is required, check for 401 status code
	if auth && resp.StatusCode == http.StatusUnauthorized {
		// Clear the token to trigger a refresh on the next request
		c.ClearToken()
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
