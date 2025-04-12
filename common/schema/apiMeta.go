//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package schema

//goland:noinspection ALL
const (
	EndpointPing             = "/api/v1/ping"
	EndpointSync             = "/api/v1/sync"
	EndpointRegister         = "/api/v1/register"
	EndpointRefresh          = "/api/v1/refresh"
	EndpointLogin            = "/api/v1/login"
	EndpointCmd              = "/api/v1/cmd"
	EndpointReport           = "/api/v1/report"
	EndpointAgent            = "/api/v1/agent"
	EndpointUser             = "/api/v1/user"
	EndpointConfigAgents     = "/api/v1/config/agent"
	EndpointConfigServer     = "/api/v1/config/server"
	EndpointReset            = "/api/v1/reset"
	EndpointRequest          = "/api/v1/request"
	EndpointRegToken         = "/api/v1/regtoken"
	EndpointEvents           = "/api/v1/events"
	EndpointCreateDeployFile = "/api/v1/deployfile"
	EndpointFiles            = "/files"
	DeployInfoFile           = "deploy.json"
)

//goland:noinspection ALL
const (
	APIStatusOK      = "ok"
	APIStatusError   = "error"
	APIStatusExpired = "expired"
)

//goland:noinspection ALL
const (
	AgentEventMessage = "message" // Notifications of events
	AgentEventAlert   = "alert"   // Alerts and warnings
	AgentEventStatus  = "status"
)

type AgentInfo struct {
	Meta   AgentMeta   `json:"meta"`
	Status AgentStatus `json:"status"`
}
