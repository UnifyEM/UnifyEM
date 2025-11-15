/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package commands

import "fmt"

type Command struct {
	Name         string   // Command name
	AckRequired  bool     // Whether the agent is expected to ack the command
	RequiredArgs []string // Required arguments
	OptionalArgs []string // Optional arguments
}

type Commands struct {
	Commands map[string]Command
}

// Command names
const (
	DownloadExecute = "download_execute"
	Execute         = "execute"
	Ping            = "ping"
	Reboot          = "reboot"
	Shutdown        = "shutdown"
	Status          = "status"
	Upgrade         = "upgrade"
	UserAdd         = "user_add"
	UserDelete      = "user_delete"
	UserAdmin       = "user_admin"
	UserPassword    = "user_password"
	UserList        = "user_list"
	UserLock        = "user_lock"
	UserUnlock      = "user_unlock"
)

// Command parameters
const (
	Arg       = "arg"
	AgentID   = "agent_id"
	RequestID = "request_id"
)

var cmds Commands

func init() {
	cmds = Commands{
		Commands: map[string]Command{
			DownloadExecute: {
				Name:         DownloadExecute,
				AckRequired:  false,
				RequiredArgs: []string{"url", "agent_id"},
				OptionalArgs: allArgN(12),
			},
			Execute: {
				Name:         Execute,
				AckRequired:  true,
				RequiredArgs: []string{"cmd", "agent_id"},
				OptionalArgs: allArgN(12),
			},
			Status: {
				Name:         Status,
				AckRequired:  true,
				RequiredArgs: []string{"agent_id"},
				OptionalArgs: []string{},
			},
			Ping: {
				Name:         Ping,
				AckRequired:  true,
				RequiredArgs: []string{"agent_id"},
				OptionalArgs: []string{},
			},
			Reboot: {
				Name:         Reboot,
				AckRequired:  false,
				RequiredArgs: []string{"agent_id"},
				OptionalArgs: []string{},
			},
			Shutdown: {
				Name:         Shutdown,
				AckRequired:  false,
				RequiredArgs: []string{"agent_id"},
				OptionalArgs: []string{},
			},
			Upgrade: {
				Name:         Upgrade,
				AckRequired:  false,
				RequiredArgs: []string{"agent_id"},
				OptionalArgs: []string{},
			},
			UserAdd: {
				Name:         UserAdd,
				AckRequired:  true,
				RequiredArgs: []string{"user", "password", "agent_id"},
				OptionalArgs: []string{"admin"},
			},
			UserDelete: {
				Name:         UserAdd,
				AckRequired:  true,
				RequiredArgs: []string{"user"},
				OptionalArgs: []string{},
			},
			UserAdmin: {
				Name:         UserAdmin,
				AckRequired:  true,
				RequiredArgs: []string{"user", "admin", "agent_id"},
				OptionalArgs: []string{},
			},
			UserPassword: {
				Name:         UserPassword,
				AckRequired:  true,
				RequiredArgs: []string{"user", "password", "agent_id"},
				OptionalArgs: []string{},
			},
			UserList: {
				Name:         UserList,
				AckRequired:  true,
				RequiredArgs: []string{"agent_id"},
				OptionalArgs: []string{},
			},
			UserLock: {
				Name:         UserLock,
				AckRequired:  true,
				RequiredArgs: []string{"user", "agent_id"},
				OptionalArgs: []string{"shutdown"},
			},
			UserUnlock: {
				Name:         UserUnlock,
				AckRequired:  true,
				RequiredArgs: []string{"user", "agent_id"},
				OptionalArgs: []string{},
			},
		},
	}

	// Add the standard parameters to all functions
	// Allows allow adding a hash. Administrators don't need to specify it, but the server will
	// add it where required
	for _, cmd := range cmds.Commands {
		cmd.OptionalArgs = append(cmd.OptionalArgs, AgentID, RequestID, "hash")
		cmds.Commands[cmd.Name] = cmd
	}
}

func allArgN(n int) []string {
	var ret []string
	for i := 1; i <= n; i++ {
		ret = append(ret, fmt.Sprintf("%s%d", Arg, i))
	}
	return ret
}

// IsAckRequired returns whether the command requires an acknowledgment from the agent
func IsAckRequired(cmd string) bool {
	command, exists := cmds.Commands[cmd]
	if !exists {
		return false
	}
	return command.AckRequired
}
