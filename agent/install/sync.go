/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package install

import (
	"fmt"
	"time"

	"github.com/UnifyEM/UnifyEM/agent/communications"
	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/agent/queues"
	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/schema/commands"
)

const (
	maxRetries    = 60 // 5 minutes with 5 second delay
	retryDelay    = 5 * time.Second
	maxRetryTotal = 5 * time.Minute
)

// syncWithRetry attempts to sync with the server with retry logic
// Retries up to 5 minutes with 5 second delay between attempts
func (i *Install) syncWithRetry(comms *communications.Communications, purpose string) error {
	startTime := time.Now()

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Check if we've exceeded total retry time
		if time.Since(startTime) > maxRetryTotal {
			return fmt.Errorf("sync for %s exceeded maximum retry time of 5 minutes", purpose)
		}

		i.logger.Infof(8130, "attempting sync for %s (attempt %d/%d)", purpose, attempt, maxRetries)

		// Attempt sync
		comms.Sync()

		// Check if sync was successful by verifying expected results
		if purpose == "registration" {
			// For registration, verify we have server public key and agent ID
			serverPublicEnc := i.config.AP.Get(global.ConfigServerPublicEnc).String()
			agentID := i.config.AP.Get(global.ConfigAgentID).String()

			if serverPublicEnc != "" && agentID != "" {
				i.logger.Infof(8131, "sync for %s successful on attempt %d", purpose, attempt)
				return nil
			}

			i.logger.Warningf(8132, "sync for %s failed on attempt %d (missing server public key or agent ID), retrying in %v",
				purpose, attempt, retryDelay)
		} else {
			// For sending credentials, verify the response queue is empty (responses were sent)
			// We can't directly check this from install, so we'll assume success after sync
			// The communications layer handles re-queuing on failure
			i.logger.Infof(8131, "sync for %s completed on attempt %d", purpose, attempt)
			return nil
		}

		// Wait before retrying
		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("sync for %s failed after %d attempts over %v", purpose, maxRetries, time.Since(startTime))
}

// sendServiceCredentialsToServer sends service credentials to the server
// This performs two syncs with retry logic:
// 1. Initial sync to register and get server public key
// 2. Second sync to send the queued credential response
func (i *Install) sendServiceCredentialsToServer() error {
	i.logger.Info(8140, "initiating service credential transmission to server", nil)

	// Verify credentials are in memory
	username, password, err := i.config.GetServiceCredentials()
	if err != nil {
		return fmt.Errorf("service credentials not in memory: %w", err)
	}

	if username == "" || password == "" {
		return fmt.Errorf("service credentials are empty")
	}

	i.logger.Info(8141, "service credentials verified in memory", nil)

	// Create queues
	requestQueue := queues.NewRequestQueue(global.TaskQueueSize)
	responseQueue := queues.NewResponseQueue(global.TaskQueueSize)

	// Create communications object
	comms, err := communications.New(
		communications.WithLogger(i.logger),
		communications.WithConfig(i.config),
		communications.WithRequestQueue(requestQueue),
		communications.WithResponseQueue(responseQueue))
	if err != nil {
		return fmt.Errorf("failed to create communications object: %w", err)
	}

	// First sync: register with server and get server public key
	i.logger.Info(8420, "syncing with server to complete registration", nil)
	err = i.syncWithRetry(comms, "registration")
	if err != nil {
		return fmt.Errorf("failed to complete registration sync: %w", err)
	}

	// Verify we have server public key
	serverPublicEnc := i.config.AP.Get(global.ConfigServerPublicEnc).String()
	if serverPublicEnc == "" {
		return fmt.Errorf("server public encryption key not received during registration")
	}

	agentID := i.config.AP.Get(global.ConfigAgentID).String()
	if agentID == "" {
		return fmt.Errorf("agent ID not received during registration")
	}

	i.logger.Info(8142, "registration successful, server public key received",
		fields.NewFields(fields.NewField("agent_id", agentID)))

	// Encrypt credentials for transmission to server
	i.logger.Info(8143, "encrypting service credentials for transmission", nil)

	// Get double-encrypted credentials directly
	encryptedForServer, err := i.config.GetServiceCredentialsForServer()
	if err != nil {
		return fmt.Errorf("failed to encrypt credentials for server: %w", err)
	}

	// Create response directly (no need for request/handler pattern for self-initiated responses)
	response := schema.NewAgentResponse()
	response.Cmd = commands.RefreshServiceAccount
	response.RequestID = "none" // Indicates unsolicited response (not in response to server request)
	response.ServiceCredentials = encryptedForServer
	response.Response = "service credentials encrypted and ready"
	response.Success = true

	// Queue the response for transmission
	responseQueue.Add(response)
	i.logger.Info(8144, "service credentials queued for transmission", nil)

	// Second sync: send the queued credentials to the server
	i.logger.Info(8421, "sending encrypted credentials to server", nil)
	err = i.syncWithRetry(comms, "credential transmission")
	if err != nil {
		return fmt.Errorf("failed to send credentials to server: %w", err)
	}

	// Verify credentials are no longer pending (they were successfully sent)
	if i.config.CredentialsPendingSend() {
		i.logger.Warning(8145, "credentials may not have been successfully transmitted (still marked as pending)", nil)
		return fmt.Errorf("credentials appear to still be pending after sync")
	}

	i.logger.Info(8146, "service credentials successfully transmitted to server", nil)

	return nil
}
