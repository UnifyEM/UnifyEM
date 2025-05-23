//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package communications

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

// Register is called from other packages when a condition such as a null agent ID is detected
func (c *Communications) Register() {
	_, err := c.register()
	if err != nil {
		c.logger.Warningf(8019, "registration failed: %s", err.Error())
		return
	}
}

// register with the UEM server
func (c *Communications) register() (string, error) {
	c.logger.Info(8015, "attempting registration", nil)

	// Get the registration key
	regToken := c.conf.AP.Get(global.ConfigRegToken).String()
	if regToken == "" {
		return "", fmt.Errorf("registration required but token is null")
	}

	// Split the token to get the server URL
	server, regToken, err := splitToken(regToken)
	if err != nil {
		return "", err
	}

	req := schema.AgentRegisterRequest{
		Token:   regToken,
		Version: global.Version,
		Build:   global.Build,
	}

	// Send the registration request
	resp, err := c.post(server, schema.EndpointRegister, false, req)
	if err != nil {
		// Check for connection error
		if strings.Contains(err.Error(), "No connection could be made") {
			return "", fmt.Errorf("unable to connect to server")
		}
		return "", err
	}

	// Unmarshal the response body into schema.ServerResponse
	var serverResponse schema.APIRegisterResponse
	err = json.Unmarshal(resp, &serverResponse)
	if err != nil {
		return "", err
	}

	if serverResponse.Code != 200 {
		return "", fmt.Errorf("registration failed with code %d", serverResponse.Code)
	}

	// Save the server URL, agent ID, and refresh token
	c.conf.AP.Set(global.ConfigServerURL, server)
	c.conf.AP.Set(global.ConfigAgentID, serverResponse.AgentID)
	c.conf.AP.Set(global.ConfigRefreshToken, serverResponse.RefreshToken)

	// Store the access token and server info locally
	c.jwt = serverResponse.AccessToken
	c.logger.Info(8016, "registration successful", fields.NewFields(fields.NewField("agent_id", serverResponse.AgentID)))

	// Checkpoint the configuration
	err = c.conf.Checkpoint()
	if err != nil {
		c.logger.Errorf(8017, "error checkpointing configuration: %s", err.Error())
	}
	return c.jwt, nil
}

// splitToken splits the token into server URL and registration token.
func splitToken(token string) (string, string, error) {

	// Parse the token as a URL
	parsedURL, err := url.Parse(token)
	if err != nil {
		return "", "", fmt.Errorf("invalid token format: %w", err)
	}

	// Ensure the URL has a scheme and host
	if parsedURL.Scheme == "" {
		return "", "", fmt.Errorf("invalid token format: missing scheme")
	}

	// Ensure the URL has a scheme and host
	if parsedURL.Host == "" {
		return "", "", fmt.Errorf("invalid token format: missing host")
	}

	// Only allow http if global.Unsafe is true
	if global.Unsafe {
		// Allow http or https
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			return "", "", fmt.Errorf("invalid token format: scheme must be HTTP or HTTPS")
		}
	} else {
		// Only allow https
		if parsedURL.Scheme != "https" {
			return "", "", fmt.Errorf("invalid token format: scheme must be HTTPS")
		}
	}

	// Extract the registration token from the path
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) != 1 {
		return "", "", fmt.Errorf("invalid token format: registration token")
	}
	regToken := pathParts[0]

	// Reconstruct the server URL without the path
	parsedURL.Path = ""
	parsedURL.RawQuery = ""
	parsedURL.Fragment = ""
	serverURL := parsedURL.String()

	return serverURL, regToken, nil
}
