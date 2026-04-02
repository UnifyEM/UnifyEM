/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package global

import (
	"encoding/json"
	"os"
	"runtime"

	"github.com/UnifyEM/UnifyEM/common/schema"
)

// UnixBackupFiles lists candidate paths for uem-backup.conf, parallel to UnixConfigFiles.
var UnixBackupFiles = []string{
	"/etc/uem-backup.conf",
	"/usr/local/etc/uem-backup.conf",
	"/var/root/uem-backup.conf",
}

// WriteBackup persists the agent's non-regeneratable identity fields to the
// first writable path in UnixBackupFiles. It is a no-op if AgentID is empty
// or on Windows (which uses the registry for persistence).
// File permissions are set to 0600 to restrict access to root/owner only.
func WriteBackup(conf *AgentConfig) error {
	if runtime.GOOS == "windows" {
		return nil
	}

	agentID := conf.AP.Get(ConfigAgentID).String()
	if agentID == "" {
		// Nothing worth backing up yet
		return nil
	}

	backup := schema.BackupConfig{
		Agent: &schema.AgentBackup{
			AgentID:         agentID,
			ServerURL:       conf.AP.Get(ConfigServerURL).String(),
			RefreshToken:    conf.AP.Get(ConfigRefreshToken).String(),
			ServerPublicSig: conf.AP.Get(ConfigServerPublicSig).String(),
			ServerPublicEnc: conf.AP.Get(ConfigServerPublicEnc).String(),
			ECPrivateSig:    conf.AP.Get(ConfigAgentECPrivateSig).String(),
			ECPublicSig:     conf.AP.Get(ConfigAgentECPublicSig).String(),
			ECPrivateEnc:    conf.AP.Get(ConfigAgentECPrivateEnc).String(),
			ECPublicEnc:     conf.AP.Get(ConfigAgentECPublicEnc).String(),
		},
	}

	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return err
	}

	var lastErr error
	for _, path := range UnixBackupFiles {
		tmp := path + ".tmp"
		if writeErr := os.WriteFile(tmp, data, 0600); writeErr != nil {
			lastErr = writeErr
			continue
		}
		if renameErr := os.Rename(tmp, path); renameErr != nil {
			_ = os.Remove(tmp)
			lastErr = renameErr
			continue
		}
		return nil
	}

	return lastErr
}

// ReadBackup reads and unmarshals the first backup file found in UnixBackupFiles.
// Returns nil, nil if no backup file exists or on Windows — not treated as an error.
// Returns an error only if a file is found but cannot be read or parsed.
func ReadBackup() (*schema.BackupConfig, error) {
	if runtime.GOOS == "windows" {
		return nil, nil
	}

	var lastErr error
	for _, path := range UnixBackupFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				lastErr = err
			}
			continue
		}

		var backup schema.BackupConfig
		if err := json.Unmarshal(data, &backup); err != nil {
			// Rename corrupt backup so it stops causing warnings on every restart
			_ = os.Rename(path, path+".bak")
			lastErr = err
			continue
		}

		return &backup, nil
	}

	return nil, lastErr
}

// RestoreFromBackup copies non-empty fields from backup.Agent into conf.AP.
// Returns true if the agent identity (AgentID) was successfully restored.
// Returns false if backup is nil, backup.Agent is nil, or AgentID is empty.
func RestoreFromBackup(conf *AgentConfig, backup *schema.BackupConfig) bool {
	if backup == nil || backup.Agent == nil || backup.Agent.AgentID == "" {
		return false
	}

	a := backup.Agent

	conf.AP.Set(ConfigAgentID, a.AgentID)

	if a.ServerURL != "" {
		conf.AP.Set(ConfigServerURL, a.ServerURL)
	}
	if a.RefreshToken != "" {
		conf.AP.Set(ConfigRefreshToken, a.RefreshToken)
	}
	if a.ServerPublicSig != "" {
		conf.AP.Set(ConfigServerPublicSig, a.ServerPublicSig)
	}
	if a.ServerPublicEnc != "" {
		conf.AP.Set(ConfigServerPublicEnc, a.ServerPublicEnc)
	}
	if a.ECPrivateSig != "" {
		conf.AP.Set(ConfigAgentECPrivateSig, a.ECPrivateSig)
	}
	if a.ECPublicSig != "" {
		conf.AP.Set(ConfigAgentECPublicSig, a.ECPublicSig)
	}
	if a.ECPrivateEnc != "" {
		conf.AP.Set(ConfigAgentECPrivateEnc, a.ECPrivateEnc)
	}
	if a.ECPublicEnc != "" {
		conf.AP.Set(ConfigAgentECPublicEnc, a.ECPublicEnc)
	}

	return true
}
