/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

// Package global provides global variables and functions for UEMAgent that can be imported by any other package
package global

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/UnifyEM/UnifyEM/common/crypto"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
	"github.com/UnifyEM/UnifyEM/common/uconfig"
)

type AgentConfig struct {
	C  interfaces.Config     // Config object
	AC interfaces.Parameters // Agent configuration
	AP interfaces.Parameters // Agent protected config

	// Non-exported fields for encrypted service credentials
	serviceCredentialsEncrypted string // Encrypted "username:password" with agent's public key
	credentialsPendingSend      bool   // True if credentials updated but not sent to server
}

// Config loads the configuration from the registry or file system
func Config() (*AgentConfig, error) {
	var err error
	c := &AgentConfig{}

	// For Windows, use the registry
	if runtime.GOOS == "windows" {
		c.C, err = uconfig.New(uconfig.WithWindowsRegistry(Name))
	} else {
		configFiles := UnixConfigFiles
		c.C, err = uconfig.New(uconfig.WithFindOrCreate(configFiles))
	}

	if err != nil {
		return &AgentConfig{}, err
	}

	// Set constraints, including default values
	c.AC, c.AP = setDefaults(c.C)

	// Update the global lost flag when the config is loaded
	Lost = c.AP.Get(ConfigLost).Bool()

	// Check for a data path
	dPath := c.AP.Get(ConfigAgentDataDir).String()
	if dPath == "" {
		var dSearch []string

		if runtime.GOOS == "windows" {
			dSearch = WindowsDefaultDataPaths
		} else {
			dSearch = UnixDefaultDataPaths
		}

		// Look for a suitable directory
		for _, path := range dSearch {
			// createDir will return true if the directory
			// exists or was successfully created
			if uconfig.CreateDir(path) {
				dPath = path
				break
			}
		}

		// Check for success
		if dPath == "" {
			return &AgentConfig{}, fmt.Errorf("unable to determine or create data directory")
		}

		// Save the path to the config
		c.AP.Set(ConfigAgentDataDir, dPath)
	}

	// Check for logfile and if not set one
	logFile := c.AP.Get(ConfigAgentLogFile).String()
	if logFile == "" {
		lPath := uconfig.CreateSubDir(dPath, "logs")
		if lPath == "" {
			// fall back to the default
			logFile = DefaultLog()
		} else {
			logFile = lPath + string(os.PathSeparator) + LogName + ".log"
		}
		c.AP.Set(ConfigAgentLogFile, logFile)
	}

	// Attempt to checkpoint the config
	_ = c.Checkpoint()
	return c, err
}

func (c *AgentConfig) Checkpoint() error {
	return c.C.Checkpoint()
}

// SetServiceCredentials encrypts and stores service account credentials in memory
// Credentials are stored as "username:password" encrypted with agent's public encryption key
// This marks credentials as pending send to server
func (c *AgentConfig) SetServiceCredentials(username, password string) error {
	// Get agent's public encryption key
	agentPublicEnc := c.AP.Get(ConfigAgentECPublicEnc).String()
	if agentPublicEnc == "" {
		return fmt.Errorf("agent public encryption key not available")
	}

	// Create plaintext "username:password"
	plaintext := username + ":" + password

	// Encrypt with agent's public key
	encrypted, err := crypto.Encrypt([]byte(plaintext), agentPublicEnc)
	if err != nil {
		return fmt.Errorf("failed to encrypt credentials: %w", err)
	}

	// Store encrypted credentials
	c.serviceCredentialsEncrypted = encrypted
	c.credentialsPendingSend = true

	return nil
}

// GetServiceCredentials decrypts and returns service account credentials
// Returns username and password as separate strings
func (c *AgentConfig) GetServiceCredentials() (username, password string, err error) {
	if c.serviceCredentialsEncrypted == "" {
		return "", "", fmt.Errorf("no credentials stored")
	}

	// Get agent's private encryption key
	agentPrivateEnc := c.AP.Get(ConfigAgentECPrivateEnc).String()
	if agentPrivateEnc == "" {
		return "", "", fmt.Errorf("agent private encryption key not available")
	}

	// Decrypt with agent's private key
	decrypted, err := crypto.Decrypt(c.serviceCredentialsEncrypted, agentPrivateEnc)
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt credentials: %w", err)
	}

	// Split "username:password"
	parts := strings.SplitN(string(decrypted), ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid credential format")
	}

	return parts[0], parts[1], nil
}

// SetServiceCredentialsEncrypted stores already-encrypted credentials received from server
// This does not mark credentials as pending send
func (c *AgentConfig) SetServiceCredentialsEncrypted(encrypted string) {
	c.serviceCredentialsEncrypted = encrypted
	c.credentialsPendingSend = false
}

// GetServiceCredentialsForServer returns credentials encrypted for transmission to server
// Credentials are double-encrypted: first with agent key (already done), then with server key
// This marks credentials as sent (no longer pending)
func (c *AgentConfig) GetServiceCredentialsForServer() (string, error) {
	if c.serviceCredentialsEncrypted == "" {
		return "", fmt.Errorf("no credentials stored")
	}

	// Get server's public encryption key
	serverPublicEnc := c.AP.Get(ConfigServerPublicEnc).String()
	if serverPublicEnc == "" {
		return "", fmt.Errorf("server public encryption key not available")
	}

	// Encrypt the already-encrypted credentials with server's public key (double encryption)
	doubleEncrypted, err := crypto.Encrypt([]byte(c.serviceCredentialsEncrypted), serverPublicEnc)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt for server: %w", err)
	}

	// Mark as sent
	c.credentialsPendingSend = false

	return doubleEncrypted, nil
}

// CredentialsPendingSend returns true if credentials need to be sent to server
func (c *AgentConfig) CredentialsPendingSend() bool {
	return c.credentialsPendingSend && c.serviceCredentialsEncrypted != ""
}
