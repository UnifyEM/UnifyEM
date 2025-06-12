//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

package display

import (
	"encoding/json"
	"fmt"

	"github.com/UnifyEM/UnifyEM/cli/credentials"

	"github.com/UnifyEM/UnifyEM/cli/global"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

func ReportResp(statusCode int, data []byte, err error) error {

	// Check for errors
	if err != nil {
		return fmt.Errorf("HTTP post failed: %w", err)
	}

	// Print the response code
	fmt.Printf("\nServer response: HTTP %d\n", statusCode)

	// Unmarshal the response body into a APICmdResponse object
	var reportResp schema.APIReportResponse
	err = json.Unmarshal(data, &reportResp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for expired access token
	if reportResp.Status == schema.APIStatusExpired {
		credentials.AccessExpired()
	}

	global.Pretty(reportResp) // TODO need a way to save a report to a file

	fmt.Println()

	if reportResp.Report.Type == schema.ReportTypeJSON {
		fmt.Println(string(reportResp.Report.Data))
		return nil
	}

	// Fallback to string format
	if len(reportResp.Report.Data) > 0 {
		if reportResp.Report.Name != "" {
			fmt.Printf("%s\n", reportResp.Report.Name)
		}

		// Convert to a string and display it
		fmt.Println(string(reportResp.Report.Data))
	}

	return nil
}
