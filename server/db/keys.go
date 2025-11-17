/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package db

import (
	"regexp"
)

// validateKey removes any invalid characters (anything other than a-z, A-Z, 0-9, -) from the input string
func validateKey(key string) string {
	validChars := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	return validChars.ReplaceAllString(key, "")
}

/*
// CombineKeys constrains both input strings to a normal character set (a-z, A-Z, 0-9, -) with no colons in them,
// and then returns "agentID:requestID" as a string
func CombineKeys(agentID, requestID string) string {
	agentID = validateKey(agentID)
	requestID = validateKey(requestID)
	return fmt.Sprintf("%s:%s", agentID, requestID)
}
*/
