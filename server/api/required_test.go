/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package api

import (
	"testing"

	"github.com/UnifyEM/UnifyEM/common/schema"
)

func TestRegisterHasRequiredFields(t *testing.T) {
	tests := []struct {
		name string
		req  schema.AgentRegisterRequest
		want bool
	}{
		{
			name: "stamped production agent",
			req:  schema.AgentRegisterRequest{Token: "tok", Version: "0.0.62", Build: 20260903},
			want: true,
		},
		{
			name: "unstamped local go build",
			req:  schema.AgentRegisterRequest{Token: "tok", Version: "0.0.62", Build: 0},
			want: true,
		},
		{
			name: "missing token",
			req:  schema.AgentRegisterRequest{Version: "0.0.62", Build: 20260903},
			want: false,
		},
		{
			name: "missing version",
			req:  schema.AgentRegisterRequest{Token: "tok", Build: 20260903},
			want: false,
		},
		{
			name: "empty unstamped",
			req:  schema.AgentRegisterRequest{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := registerHasRequiredFields(tt.req); got != tt.want {
				t.Fatalf("registerHasRequiredFields(%+v) = %v, want %v", tt.req, got, tt.want)
			}
		})
	}
}

func TestSyncHasRequiredFields(t *testing.T) {
	tests := []struct {
		name string
		req  schema.AgentSyncRequest
		want bool
	}{
		{
			name: "stamped production agent",
			req:  schema.AgentSyncRequest{Version: "0.0.62", Build: 20260903},
			want: true,
		},
		{
			name: "unstamped local go build",
			req:  schema.AgentSyncRequest{Version: "0.0.62", Build: 0},
			want: true,
		},
		{
			name: "missing version",
			req:  schema.AgentSyncRequest{Build: 20260903},
			want: false,
		},
		{
			name: "empty unstamped",
			req:  schema.AgentSyncRequest{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := syncHasRequiredFields(tt.req); got != tt.want {
				t.Fatalf("syncHasRequiredFields(%+v) = %v, want %v", tt.req, got, tt.want)
			}
		})
	}
}
