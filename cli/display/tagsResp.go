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

// TagsResp handles schema.APIRequestStatusResponse from the server.
func TagsResp(statusCode int, data []byte, err error) error {

	// Check for errors
	if err != nil {
		return fmt.Errorf("HTTP post failed: %w", err)
	}

	// Print the response code
	fmt.Printf("\nServer response: HTTP %d\n", statusCode)

	// Unmarshal the response body into the correct object
	var resp schema.AgentTagsResponse
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
