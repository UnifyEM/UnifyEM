/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package api

import "github.com/UnifyEM/UnifyEM/common/schema"

// registerHasRequiredFields reports whether a registration request has the
// fields the server needs. Build may be 0: that is app.BuildDate() for an
// unstamped binary (plain go build / go run / IDE).
func registerHasRequiredFields(req schema.AgentRegisterRequest) bool {
	return req.Token != "" && req.Version != ""
}

// syncHasRequiredFields reports whether a sync request has the fields the
// server needs. Build may be 0 for the same reason as registration.
func syncHasRequiredFields(req schema.AgentSyncRequest) bool {
	return req.Version != ""
}
