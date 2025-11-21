/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package install

import (
	"fmt"
	"time"

	"github.com/UnifyEM/UnifyEM/agent/communications"
	"github.com/UnifyEM/UnifyEM/agent/functions"
	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/agent/queues"
	"github.com/UnifyEM/UnifyEM/common/schema"
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
	fmt.Println("Syncing with server to complete registration...")
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

	i.logger.Info(8142, "registration successful, server public key received", nil)
	fmt.Printf("Registration successful. Agent ID: %s\n", agentID)

	// Create functions handler to process update_service_account request
	cmd, err := functions.New(
		functions.WithLogger(i.logger),
		functions.WithConfig(i.config),
		functions.WithComms(comms),
		functions.WithUserDataSource(nil)) // user data source not needed for this command
	if err != nil {
		return fmt.Errorf("failed to initialize command module: %w", err)
	}

	// Create the update_service_account request
	request := schema.NewAgentRequest()
	request.AgentID = agentID
	request.RequestID = "update_service_account"
	request.Request = "update_service_account"
	request.Parameters = make(map[string]string)
	request.Parameters["agent_id"] = agentID

	// Execute the request to queue the encrypted credentials response
	i.logger.Info(8143, "encrypting service credentials for transmission", nil)
	fmt.Println("Encrypting service credentials for transmission...")

	response := cmd.ExecuteRequest(request)

	if !response.Success {
		return fmt.Errorf("update_service_account command failed: %s", response.Response)
	}

	// Add response to queue
	responseQueue.Add(response)
	i.logger.Info(8144, "service credentials queued for transmission", nil)

	// Second sync: send the queued credentials to the server
	fmt.Println("Sending encrypted credentials to server...")
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
	fmt.Println("Service credentials successfully transmitted to server.")

	return nil
}
