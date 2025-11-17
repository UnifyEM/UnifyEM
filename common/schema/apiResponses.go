/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package schema

import (
	"time"
)

// By design, there is a high level of redundancy in the API response structures. However, more specific
// structures make the API easier to understand and provide relevant examples for swaggo to generate swagger
// documentation.

// All API responses must include the Status and Code fields.

// GeneralAPIResponse can be used by a client to deserialize any API response. However, when data is included,
// a more specific struct may make it easier for the client.

type APIAnyResponse struct {
	Status       string `json:"status"`                  // API status response - see schema/apiMeta.go
	Code         int    `json:"code"`                    // HTTP status code
	Details      string `json:"details,omitempty"`       // Optional details about the response
	AgentID      string `json:"agent_id,omitempty"`      // Agent ID during registration
	RequestID    string `json:"request_id,omitempty"`    // Unique identifier for a request (command) sent to an agent
	AccessToken  string `json:"access_token,omitempty"`  // JWT access token during registration, authentication, and refresh
	RefreshToken string `json:"refresh_token,omitempty"` // JWT refresh token during registration and authentication
	Report       Report `json:"report,omitempty"`        // Report if applicable
	Data         any    `json:"data,omitempty"`          // optional data
}

// All API4xx and 5xx responses have the same structure. This is solely to assist swaggo generate docs.

type API400 struct {
	Status  string `json:"status" example:"error"`
	Code    int    `json:"code" example:"400"`
	Details string `json:"details" example:"bad request"`
}

type API401 struct {
	Status  string `json:"status" example:"error"`
	Code    int    `json:"code" example:"401"`
	Details string `json:"details" example:"authentication failed"`
}

type API404 struct {
	Status  string `json:"status" example:"error"`
	Code    int    `json:"code" example:"404"`
	Details string `json:"details" example:"object not found"`
}

type API500 struct {
	Status  string `json:"status" example:"error"`
	Code    int    `json:"code" example:"500"`
	Details string `json:"details" example:"internal server error"`
}

// APIGenericResponse is used for successful responses that don't require a specific structure
type APIGenericResponse struct {
	Status  string `json:"status" example:"ok"`                           // API status response - see schema/apiMeta.go
	Code    int    `json:"code" example:"200"`                            // HTTP status code
	Details string `json:"details,omitempty" example:"request processed"` // Optional response details
}

type APILoginResponse struct {
	Status       string `json:"status" example:"ok"`                   // Text Status
	Code         int    `json:"code" example:"200"`                    // HTTP status code
	AccessToken  string `json:"access_token,omitempty" example:"jwt"`  // JWT access token
	RefreshToken string `json:"refresh_token,omitempty" example:"jwt"` // JTW refresh token
}

type APITokenRefreshResponse struct {
	Status      string `json:"status" example:"ok"`                  // Text Status
	Code        int    `json:"code" example:"200"`                   // HTTP status code
	AccessToken string `json:"access_token,omitempty" example:"jwt"` // JWT access token
}

type APIAgentInfoResponse struct {
	Status  string    `json:"status" example:"ok"`
	Code    int       `json:"code" example:"200"`
	Details string    `json:"details,omitempty" example:"agent info"`
	Data    AgentList `json:"data"`
}

type APIEventsResponse struct {
	Status  string       `json:"status" example:"ok"`
	Code    int          `json:"code" example:"200"`
	Details string       `json:"details,omitempty" example:"events"`
	Data    []AgentEvent `json:"data"`
}

type AgentEvent struct {
	AgentID   string            `json:"agent_id"`
	EventID   string            `json:"event_id"`
	Time      time.Time         `json:"time"`
	EventType string            `json:"type"`
	Event     string            `json:"event"`
	Details   map[string]string `json:"details,omitempty"`
}

type APIConfigResponse struct {
	Status  string            `json:"status" example:"ok"`
	Code    int               `json:"code" example:"200"`
	Details string            `json:"details,omitempty" example:"Config set retrieved"`
	Data    map[string]string `json:"data"`
}

// APISyncResponse is sent to the agent by the server in response to an agent sync request
type APISyncResponse struct {
	Status   string            `json:"status"`
	Code     int               `json:"code"`
	Conf     map[string]string `json:"conf"`
	Triggers AgentTriggers     `json:"triggers"`
	Details  string            `json:"details,omitempty"`
	Requests []AgentRequest    `json:"requests"` // Requests for the agent to process and respond to
}

// AgentRequest contains a single command (request) from the server to the agent
type AgentRequest struct {
	Created     time.Time         `json:"created"`
	Requester   string            `json:"requester"`
	Request     string            `json:"request"`
	AckRequired bool              `json:"ack_required"`
	RequestID   string            `json:"request_id"`
	AgentID     string            `json:"agent_id"`
	Parameters  map[string]string `json:"parameters"`
}

// NewAgentRequest creates a new AgentRequest and initializes the map to avoid errors
func NewAgentRequest() AgentRequest {
	return AgentRequest{
		Parameters: make(map[string]string),
	}
}

// APIReportResponse is used by the API to respond to a report request
type APIReportResponse struct {
	Status  string `json:"status"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
	Report  Report `json:"report"`
}

// APIRequestStatusResponse contains information about one or more request (command) sent to an agent
type APIRequestStatusResponse struct {
	Status  string                 `json:"status"`
	Code    int                    `json:"code"`
	Details string                 `json:"details,omitempty"`
	Data    AgentRequestRecordList `json:"report"`
}

// APIRegisterResponse is sent to the agent by the server in response to a registration agent
type APIRegisterResponse struct {
	Status       string `json:"status"`
	Code         int    `json:"code"`
	Details      string `json:"details,omitempty"`
	AgentID      string `json:"agent_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// APICmdResponse is used by the API to respond to a command request
type APICmdResponse struct {
	Status            string `json:"status" example:"ok"`
	Code              int    `json:"code" example:"200"`
	Details           string `json:"details,omitempty" example:"request queued for agent"`
	RequestID         string `json:"request_id,omitempty" example:"R-6f9dcb2e-2e1b-4c3a-8a67-5b3e0d740df6"`
	AgentID           string `json:"agent_id,omitempty" example:"A-12345678-abcd-1234-5648-1234567890ab"`
	AgentFriendlyName string `json:"agent_friendly_name,omitempty" example:"Tuxedo001 Linux Laptop"`
}
