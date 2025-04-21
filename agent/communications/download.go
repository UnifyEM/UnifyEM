//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package communications

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/UnifyEM/UnifyEM/agent/global"
)

func (c *Communications) Download(url string) (string, error) {
	var err error

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", global.Name+"-tmp-*")
	if err != nil {
		return "", fmt.Errorf("error creating temporary file: %w", err)
	}

	// Get the server URL
	server, err := getServerURL(url)
	if err != nil {
		return "", fmt.Errorf("error parsing URL: %w", err)
	}

	// Get the server URL
	ourServer := c.conf.AP.Get(global.ConfigServerURL).String()
	if ourServer == "" {
		return "", fmt.Errorf("unable to obtain ServerURL %w", err)
	}

	// Only send authentication to our server
	var token = ""
	if server == ourServer {
		// If authentication is required, obtain and set the bearer token
		// GetToken() will attempt refresh or registration if required
		token, err = c.GetToken()
		if err != nil {
			return "", err
		}
	} else {
		if global.PROTECTED {
			return "", fmt.Errorf("protected mode on, refusing to download %s", url)
		}
	}

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating http request: %w", err)
	}

	// If token is not empty, set the Authorization header
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	// Create an HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending http request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Check the HTTP status code
	if resp.StatusCode != http.StatusOK {

		// Close and delete the temporary file
		closeDelete(tmpFile)
		return "", fmt.Errorf("error downloading %s: received status code %d", url, resp.StatusCode)
	}

	// Write the downloaded content to the temporary file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		closeDelete(tmpFile)
		return "", fmt.Errorf("error writing to temporary file: %w", err)
	}

	// Close the temporary file
	err = tmpFile.Close()
	if err != nil {
		closeDelete(tmpFile)
		return "", fmt.Errorf("error closing temporary file: %w", err)
	}

	// Return the temporary file name
	return tmpFile.Name(), nil
}

// closeDelete closes the file, deletes it, and ignores any errors
func closeDelete(file *os.File) {
	if file == nil {
		return
	}

	_ = file.Close()
	_ = os.Remove(file.Name())
}

// getServerURL extracts the scheme, server, and port from the given URL
func getServerURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %w", err)
	}

	// Reconstruct the server URL with scheme, host, and port
	serverURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	return serverURL, nil
}
