//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package schema

import "time"

type AgentMeta struct {
	AgentID      string        `json:"agent_id"`
	Active       bool          `json:"active"`
	FriendlyName string        `json:"friendly_name"`
	FirstSeen    time.Time     `json:"first_seen"`
	LastSeen     time.Time     `json:"last_seen"`
	LastIP       string        `json:"last_ip"`
	Version      string        `json:"version"`
	Build        int           `json:"build"`
	Triggers     AgentTriggers `json:"triggers"`
	Status       *AgentStatus  `json:"status,omitempty"`
	Tags         []string      `json:"tags"`
}

func NewAgentMeta(agentID string) AgentMeta {
	now := time.Now()
	return AgentMeta{
		AgentID:      agentID,
		Active:       true,
		FriendlyName: "",
		FirstSeen:    now,
		LastSeen:     now,
		Triggers:     NewAgentTriggers(),
		Status:       nil,
		Tags:         []string{},
	}
}

type AgentStatus struct {
	LastUpdated time.Time         `json:"last_updated"`
	Details     map[string]string `json:"details"`
}

// Request for adding/removing tags
type AgentTagsRequest struct {
	Tags []string `json:"tags"`
}

// Response for tag operations
type AgentTagsResponse struct {
	Tags   []string `json:"tags"`
	Status string   `json:"status"`
	Code   int      `json:"code"`
}

// Response for agents by tag
type AgentsByTagResponse struct {
	Agents []AgentMeta `json:"agents"`
	Status string      `json:"status"`
	Code   int         `json:"code"`
}

// AgentTriggers are used to request specific immediate actions from the agent
// If a new trigger is added, it should be added to the NewAgentTriggers function
// to ensure a proper reset is possible
type AgentTriggers struct {
	Lost      bool `json:"lost" example:"false"`
	Uninstall bool `json:"uninstall" example:"false"`
	Wipe      bool `json:"wipe" example:"false"`
}

// NewAgentTriggers creates a new AgentTriggers struct with all values set to false
func NewAgentTriggers() AgentTriggers {
	return AgentTriggers{Lost: false, Uninstall: false, Wipe: false}
}

type AgentList struct {
	Agents []AgentMeta `json:"agents"`
}

// DeviceUserList is used to return a list of pc/device users
type DeviceUserList struct {
	Users []DeviceUser
}

type DeviceUser struct {
	Domain        string
	Name          string
	Disabled      bool
	Administrator bool
}
