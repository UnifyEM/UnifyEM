/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package communications

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

// GetToken returns the token, refreshing if required
func (c *Communications) GetToken() (string, error) {
	var err error

	if c.jwt == "" {
		// Refresh the token
		c.jwt, err = c.refreshToken()
		if err != nil {
			return "", err
		}
	}
	return c.jwt, nil
}

// ClearToken is used to clear the token when an auth failure occurs
func (c *Communications) ClearToken() {
	// Clear the token
	c.jwt = ""
	c.retryRequired = true
}

// refreshToken will attempt to refresh the token and will register if necessary
func (c *Communications) refreshToken() (string, error) {
	var rToken string

	// Get the Refresh Token
	rToken = c.conf.AP.Get(global.ConfigRefreshToken).String()
	if rToken != "" {

		c.logger.Info(8010, "attempting access token refresh", nil)

		// Get the server URL
		serverURL := c.conf.AP.Get(global.ConfigServerURL).String()
		if serverURL == "" {
			return "", errors.New("error fetching ServerURL")
		}

		// Get agent's public keys from configuration (for rekey scenarios)
		clientPublicSig := c.conf.AP.Get(global.ConfigAgentECPublicSig).String()
		clientPublicEnc := c.conf.AP.Get(global.ConfigAgentECPublicEnc).String()

		// Create a refresh request to the server
		req := schema.RefreshRequest{
			RefreshToken:    rToken,
			ClientPublicSig: clientPublicSig,
			ClientPublicEnc: clientPublicEnc,
		}

		// Post the refresh request to the server
		data, err := c.post(serverURL, schema.EndpointRefresh, false, req)
		if err != nil {
			return "", fmt.Errorf("token refresh failed: %w", err)
		}

		// Unmarshal the response body into a LoginResponse object
		var refreshResponse schema.APITokenRefreshResponse
		err = json.Unmarshal(data, &refreshResponse)
		if err != nil {
			return "", fmt.Errorf("deserialization failed %w", err)
		}

		if refreshResponse.Code != 200 {
			c.logger.Errorf(8011, "token refresh failed with code %d", refreshResponse.Code)

			// Attempt re-registration
			token, rErr := c.register()
			if rErr != nil {
				c.logger.Errorf(8012, "registration failed: %s", rErr.Error())
			}
			return token, rErr
		}

		// Check for server public key changes (should never happen - indicates potential security issue)
		if refreshResponse.ServerPublicSig != "" {
			existing := c.conf.AP.Get(global.ConfigServerPublicSig).String()
			if existing == "" {
				// No existing key - store it
				c.conf.AP.Set(global.ConfigServerPublicSig, refreshResponse.ServerPublicSig)
				c.logger.Info(8020, "server public signature key received and stored", nil)
			} else if existing != refreshResponse.ServerPublicSig {
				// Key changed - security warning, do NOT update
				c.logger.Warning(8022, "different server public signature key received and ignored (possible security issue)", nil)
			}
		}
		if refreshResponse.ServerPublicEnc != "" {
			existing := c.conf.AP.Get(global.ConfigServerPublicEnc).String()
			if existing == "" {
				// No existing key - store it
				c.conf.AP.Set(global.ConfigServerPublicEnc, refreshResponse.ServerPublicEnc)
				c.logger.Info(8021, "server public encryption key received and stored", nil)
			} else if existing != refreshResponse.ServerPublicEnc {
				// Key changed - security warning, do NOT update
				c.logger.Warning(8023, "different server public encryption key received and ignored (possible security issue)", nil)
			}
		}

		c.logger.Info(8013, "access token refresh successful", nil)
		return refreshResponse.AccessToken, nil

	}

	// Registration is required
	token, rErr := c.register()
	if rErr != nil {
		c.logger.Errorf(8014, "registration failed: %s", rErr.Error())
	}
	return token, rErr
}
