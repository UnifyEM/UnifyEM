//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package userver

import (
	"net"
	"net/http"
	"strings"
)

// RemoteIP returns the remote IP address from the agent, excluding the port number.
//
//goland:noinspection GoUnusedExportedFunction
func RemoteIP(req *http.Request) string {

	// Check for the X-Forwarded-For header first
	forwarded := req.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// The X-Forwarded-For header can contain multiple IPs, take the first one
		ip := strings.Split(forwarded, ",")[0]
		return strings.TrimSpace(ip)
	}

	// Fallback to using the remote address from the agent
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr // Return the full address if splitting fails
	}
	return ip
}
