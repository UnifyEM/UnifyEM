//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

// Linux specific functions
//go:build linux

package privcheck

import (
	"os"
)

func Check() (bool, error) {
	if os.Geteuid() == 0 {
		return true, nil
	}
	return false, nil
}
