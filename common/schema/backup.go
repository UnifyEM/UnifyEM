/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package schema

// BackupConfig is the top-level structure for uem-backup.conf.
// Each sub-struct is a pointer so that omitempty suppresses absent sections.
// This struct is intentionally extensible — add new sub-structs here as
// additional components (CLI, server) require backup support.
type BackupConfig struct {
	Agent  *AgentBackup  `json:"agent,omitempty"`
	CLI    *CLIBackup    `json:"cli,omitempty"`    // reserved for future CLI recovery data
	Server *ServerBackup `json:"server,omitempty"` // reserved for future server recovery data
}

// AgentBackup holds the minimum fields required to recover an agent's
// identity and re-authenticate with the server without re-registering.
//
// TODO: Add any additional non-regeneratable keys (e.g., hardware-bound
// keys, CA hashes, service account credentials) as they are introduced.
type AgentBackup struct {
	AgentID         string `json:"agent_id,omitempty"`
	ServerURL       string `json:"server_url,omitempty"`
	RefreshToken    string `json:"refresh_token,omitempty"`
	ServerPublicSig string `json:"server_public_sig,omitempty"`
	ServerPublicEnc string `json:"server_public_enc,omitempty"`
	ECPrivateSig    string `json:"ec_private_sig,omitempty"`
	ECPublicSig     string `json:"ec_public_sig,omitempty"`
	ECPrivateEnc    string `json:"ec_private_enc,omitempty"`
	ECPublicEnc     string `json:"ec_public_enc,omitempty"`
}

// CLIBackup is reserved for future CLI recovery data.
type CLIBackup struct{}

// ServerBackup is reserved for future server recovery data.
type ServerBackup struct{}
