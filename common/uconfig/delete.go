/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package uconfig

import (
	"os"
)

// / Delete the configuration file
func (c *UConfig) deleteFile() error {
	return os.Remove(c.file)
}
