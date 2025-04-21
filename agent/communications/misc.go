//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package communications

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/UnifyEM/UnifyEM/agent/global"
)

// buildURL constructs a full URL by appending the path to the server URL
// It performs some basic validation to ensure the server URL doesn't include a path, etc.
func buildURL(server, path string) (string, error) {
	// Parse the server URL
	parsedURL, err := url.Parse(server)
	if err != nil {
		return "", err
	}

	// Ensure the server URL does not have anything beyond the optional port
	parsedURL.Path = ""
	parsedURL.RawQuery = ""
	parsedURL.Fragment = ""

	// Only allow http if global.Unsafe is true
	if global.Unsafe {
		// Allow http or https
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			return "", fmt.Errorf("server URL must use HTTP or HTTPS")
		}
	} else {
		// Only allow https
		if parsedURL.Scheme != "https" {
			return "", fmt.Errorf("server URL must use HTTPS")
		}
	}

	// Append the path to the server URL
	return fmt.Sprintf("%s%s", strings.TrimRight(parsedURL.String(), "/"), path), nil
}
