//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

// MacOS (Darin) specific functions
//go:build darwin

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
