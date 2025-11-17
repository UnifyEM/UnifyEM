/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package communications

import "github.com/UnifyEM/UnifyEM/cli/global"

// Ensure that Communications implements the global.Comms interface
var _ global.Comms = &Communications{}

type Communications struct {
	token string
}

// New returns a new Communications object and optionally accepts a token
func New(token ...string) global.Comms {
	comms := &Communications{
		token: "",
	}
	if len(token) > 0 {
		comms.token = token[0]
	}
	return comms
}

func (c *Communications) SetToken(token string) {
	c.token = token
}
