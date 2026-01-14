//go:build linux

/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package runCmd

import "fmt"

// SSH is not implemented on Linux
func (r *Runner) SSH(user *UserLogin, cmdAndArgs ...string) (string, error) {
	return "", fmt.Errorf("SSH execution not implemented on Linux")
}
