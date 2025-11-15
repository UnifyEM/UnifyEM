/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

// Code for operating systems other than windows
//go:build linux || darwin

package uconfig

import (
	"errors"
)

func (c *UConfig) saveRegistry() error {
	return errors.New("registry not supported on this platform")
}

func (c *UConfig) loadRegistry() error {
	return errors.New("registry not supported on this platform")
}
