/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package login

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"

	"github.com/UnifyEM/UnifyEM/cli/communications"
	"github.com/UnifyEM/UnifyEM/cli/credentials"
	"github.com/UnifyEM/UnifyEM/cli/global"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

// Login does its own error handling to avoid a lot of duplication
func Login() string {

	// If we already have an access token, return it
	accessToken := credentials.GetAccessToken()
	if accessToken != "" {
		return accessToken
	}

	// If we have a refresh token, try to refresh the access token
	refreshToken := credentials.GetRefreshToken()
	if refreshToken != "" {
		token := RefreshToken(refreshToken)
		if token != "" {
			credentials.SetAccessToken(token)
			return token
		} else {
			// Refresh failed, so we need to log in again
			credentials.RefreshExpired()
		}
	}

	// Get the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fatal(err)
	}

	// Construct the full path to the .openuem file
	envPath := filepath.Join(homeDir, ".uem")

	// Load environment variables from .openuem file if it exists
	_ = godotenv.Load(envPath)

	// Read from environment variables
	user := os.Getenv("UEM_USER")
	pass := os.Getenv("UEM_PASS")
	global.ServerURL = os.Getenv("UEM_SERVER")

	if user == "" {
		fatal(errors.New("UEM_USER is not set"))
	}

	if pass == "" {
		fatal(errors.New("UEM_PASS is not set"))
	}

	if global.ServerURL == "" {
		fatal(errors.New("UEM_SERVER is not set"))
	}

	// Create a login request
	req := schema.NewLoginRequest(user, pass)

	// Post the login request to the server
	c := communications.New()
	code, data, err := c.Post(schema.EndpointLogin, req)
	if err != nil {
		fatal(err)
	}

	if code != 200 {
		fatal(fmt.Errorf("login failed with HTTP status %d", code))
	}

	// Unmarshal the response body into a LoginResponse object
	var loginResp schema.APILoginResponse
	err = json.Unmarshal(data, &loginResp)
	if err != nil {
		fatal(fmt.Errorf("failed to unmarshal response: %w", err))
	}

	if loginResp.AccessToken == "" || loginResp.RefreshToken == "" {
		fatal(errors.New("server returned an empty token"))
	}

	// Save the tokens
	credentials.SetAccessToken(loginResp.AccessToken)
	credentials.SetRefreshToken(loginResp.RefreshToken)
	return loginResp.AccessToken
}

func fatal(err error) {
	fmt.Printf("Error: %s\n\n", err.Error())
	os.Exit(1)
}

func RefreshToken(rToken string) string {

	// Send a refresh request to the server
	req := schema.RefreshRequest{RefreshToken: rToken}

	// Post the refresh request to the server
	c := communications.New()
	code, data, err := c.Post(schema.EndpointRefresh, req)
	if err != nil {
		fmt.Printf("Token refresh failed: %s\n", err.Error())
		return ""
	}

	if code != 200 {
		// The refresh token was invalid or another error occurred
		// Force the user to log in again
		return ""
	}

	// Unmarshal the response body into a LoginResponse object
	var loginResp schema.APITokenRefreshResponse
	err = json.Unmarshal(data, &loginResp)
	if err != nil {
		fatal(fmt.Errorf("deserialization failed %w", err))
	}

	return loginResp.AccessToken
}
