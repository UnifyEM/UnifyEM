/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package display

import (
	"encoding/json"
	"fmt"

	"github.com/UnifyEM/UnifyEM/cli/credentials"
	"github.com/UnifyEM/UnifyEM/cli/global"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

// AnyResp handles any response from the server and pretty-prints it to stdout
// It also checks for an expired access token
func AnyResp(statusCode int, data []byte, err error) error {

	// Check for errors
	if err != nil {
		return fmt.Errorf("HTTP post failed: %w", err)
	}

	// Print the response code
	fmt.Printf("\nServer response: HTTP %d\n", statusCode)

	// Unmarshal the response body into a generic response object
	var resp schema.APIAnyResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for expired access token
	if resp.Status == schema.APIStatusExpired {
		credentials.AccessExpired()
	}

	global.Pretty(resp)
	return nil
}
