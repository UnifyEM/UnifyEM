/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package communications

import "testing"

func TestHostFromURL(t *testing.T) {
	tests := []struct {
		name     string
		rawURL   string
		expected string
	}{
		{
			name:     "https with explicit port",
			rawURL:   "https://server.example.com:8443/api/v1/ping",
			expected: "server.example.com:8443",
		},
		{
			name:     "https without port",
			rawURL:   "https://server.example.com/api/v1/ping",
			expected: "server.example.com:443",
		},
		{
			name:     "http without port",
			rawURL:   "http://server.example.com/api/v1/ping",
			expected: "server.example.com:80",
		},
		{
			name:     "http with explicit port",
			rawURL:   "http://server.example.com:9090/path",
			expected: "server.example.com:9090",
		},
		{
			name:     "localhost with port",
			rawURL:   "https://localhost:443/test",
			expected: "localhost:443",
		},
		{
			name:     "IP address with port",
			rawURL:   "https://192.168.1.1:8443/api",
			expected: "192.168.1.1:8443",
		},
		{
			name:     "IP address without port",
			rawURL:   "https://192.168.1.1/api",
			expected: "192.168.1.1:443",
		},
		{
			name:     "no path",
			rawURL:   "https://example.com",
			expected: "example.com:443",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := hostFromURL(tc.rawURL)
			if got != tc.expected {
				t.Errorf("hostFromURL(%q) = %q, want %q", tc.rawURL, got, tc.expected)
			}
		})
	}
}
