/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package common

import (
	"strings"
)

// SingleLine normalizes a string for logging:
//   - trims leading/trailing whitespace
//   - replaces newlines with a visible marker
//   - collapses runs of whitespace into single spaces
func SingleLine(s string) string {
	if s == "" {
		return s
	}

	// Trim outer whitespace first
	s = strings.TrimSpace(s)

	// Replace CRLF / LF / CR with a marker
	// Change " ⏎ " to whatever you like: " \\n ", " | ", " [NL] ", etc.
	replacer := strings.NewReplacer(
		"\r\n", " ⏎ ",
		"\n", " ⏎ ",
		"\r", " ⏎ ",
	)

	s = replacer.Replace(s)

	// Optional: collapse any weird spacing into single spaces
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}
