/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/UnifyEM/UnifyEM/cli/display"
	"github.com/UnifyEM/UnifyEM/cli/global"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

// waitForResponses polls the server for request status until all requests complete or a timeout occurs
func waitForResponses(c global.Comms, requestIDs []string, timeout int) error {
	if len(requestIDs) == 0 {
		return nil
	}

	// Create a map of pending requests
	pendingRequests := make(map[string]bool)
	for _, id := range requestIDs {
		pendingRequests[id] = true
	}

	fmt.Printf("\nWaiting for response(s) (timeout: %ds)...\n", timeout)
	startTime := time.Now()

	// Poll until all requests are complete or timeout
	for len(pendingRequests) > 0 {

		// Poll each pending request and collect completed ones
		completedRequests := make([]string, 0)
		for requestID := range pendingRequests {
			if checkAndDisplayIfComplete(c, requestID) {
				completedRequests = append(completedRequests, requestID)
			}
		}

		// Remove completed requests from pending map (safe to do after iteration)
		for _, requestID := range completedRequests {
			delete(pendingRequests, requestID)
		}

		// Check timeout
		elapsed := int(time.Since(startTime).Seconds())
		if elapsed >= timeout {
			displayTimeoutMessage(c, pendingRequests, elapsed)
			return nil
		}

		// Sleep before next poll cycle (only if there are still pending requests)
		if len(pendingRequests) > 0 {
			time.Sleep(5 * time.Second)
		}
	}

	return nil
}

// checkAndDisplayIfComplete polls a single request and displays it if complete
// Returns true if the request is complete, false otherwise
func checkAndDisplayIfComplete(c global.Comms, requestID string) bool {
	statusCode, data, err := c.Get(schema.EndpointRequest + "/" + requestID)
	if err != nil {
		// Network error - keep polling
		return false
	}

	if statusCode != 200 {
		// Request not found or error - consider it complete to remove from pending
		return true
	}

	// Parse response
	var resp schema.APIRequestStatusResponse
	if err = json.Unmarshal(data, &resp); err != nil {
		// Parse error - keep polling
		return false
	}

	// Check if request has completed
	if len(resp.Data.Requests) > 0 {
		request := resp.Data.Requests[0]
		if isRequestComplete(request.Status) {
			// Display the completed request
			fmt.Printf("\n")
			display.ErrorWrapper(display.RequestList(statusCode, data, nil))
			return true
		}
	}

	return false
}

// displayTimeoutMessage shows timeout information and lists non-responsive agents
func displayTimeoutMessage(c global.Comms, pendingRequests map[string]bool, elapsed int) {

	// Collect agent IDs from pending requests
	var nonResponsiveAgents []string
	for requestID := range pendingRequests {
		statusCode, data, err := c.Get(schema.EndpointRequest + "/" + requestID)
		if err == nil && statusCode == 200 {
			var resp schema.APIRequestStatusResponse
			if err := json.Unmarshal(data, &resp); err == nil {
				if len(resp.Data.Requests) > 0 {
					nonResponsiveAgents = append(nonResponsiveAgents, resp.Data.Requests[0].AgentID)
				}
			}
		}
	}

	// Display timeout message with non-responsive agents
	fmt.Printf("\n")
	fmt.Printf("Wait timed out after %ds\n", elapsed)
	if len(nonResponsiveAgents) > 0 {
		fmt.Printf("The following agent(s) have not responded yet:\n")
		for _, agentID := range nonResponsiveAgents {
			fmt.Printf("  - %s\n", agentID)
		}
	}
}

/*
// displayRequestStatus shows the current status of a request
func displayRequestStatus(c global.Comms, requestID string) {
	statusCode, data, err := c.Get(schema.EndpointRequest + "/" + requestID)
	if err == nil {
		fmt.Printf("\n")
		display.ErrorWrapper(display.RequestList(statusCode, data, nil))
	} else {
		fmt.Printf("\nUnable to retrieve status for request %s: %v\n", requestID, err)
	}
}
*/

// isRequestComplete checks if a request status indicates completion
func isRequestComplete(status string) bool {
	return status == schema.RequestStatusComplete ||
		status == schema.RequestStatusFailed ||
		status == schema.RequestStatusInvalid
}
