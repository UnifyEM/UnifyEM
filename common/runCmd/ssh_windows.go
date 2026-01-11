//go:build windows

/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package runCmd

import "fmt"

// SSH is not implemented on Windows
func (r *Runner) SSH(user *UserLogin, cmdAndArgs ...string) (string, error) {
	return "", fmt.Errorf("SSH execution not implemented on Windows")
}
