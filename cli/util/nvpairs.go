//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

package util

import "strings"

type NVPairs struct {
	Pairs map[string]string
}

// NewNVPairs parses a list of strings for key=value pairs and returns them in a map
func NewNVPairs(args []string) *NVPairs {
	r := NVPairs{
		Pairs: make(map[string]string),
	}

	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			r.Pairs[strings.ToLower(parts[0])] = parts[1]
		}
	}

	return &r
}

// ToMap is a helper function to convert NVPairs to a map[string]string
func (p *NVPairs) ToMap() map[string]string {
	return p.Pairs
}
