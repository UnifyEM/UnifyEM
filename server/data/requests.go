//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package data

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/schema/commands"
	"github.com/UnifyEM/UnifyEM/server/global"
)

// GetAgentRequest returns a single request for an agent
func (d *Data) GetAgentRequest(requestKey string) (schema.AgentRequest, error) {
	request, err := d.database.GetAgentRequest(requestKey)
	if err != nil {
		return schema.AgentRequest{}, fmt.Errorf("error getting agent request: %w", err)
	}

	return schema.AgentRequest{
		Requester:  request.Requester,
		RequestID:  request.RequestID,
		Request:    request.Request,
		Parameters: request.Parameters,
	}, nil
}

// GetRequestRecord returns the entire record for an administrator
// A list is used for consistency with GetAllRequestRecords
func (d *Data) GetRequestRecord(requestKey string) (schema.AgentRequestRecordList, error) {
	request, err := d.database.GetAgentRequest(requestKey)
	if err != nil {
		return schema.AgentRequestRecordList{}, fmt.Errorf("error getting agent request: %w", err)
	}

	return schema.AgentRequestRecordList{Requests: []schema.AgentRequestRecord{request}}, nil
}

func (d *Data) GetRequestRecords() (schema.AgentRequestRecordList, error) {
	return d.database.GetAllRequestRecords()
}

// DeleteAgentRequest removes a request from the database
func (d *Data) DeleteAgentRequest(requestKey string) error {
	return d.database.DeleteAgentRequest(requestKey)
}

// GetAgentRequests returns a list of requests for an agent
// If markSent is true, commands that do not require an ack are marked complete
func (d *Data) GetAgentRequests(agentID string, markSent bool) ([]schema.AgentRequest, error) {
	var requestList []schema.AgentRequest

	// Get a list of requests for this agent
	requests, err := d.database.GetAgentRequests(agentID)
	if err != nil {
		d.logger.Error(2704, "error getting agent requests",
			fields.NewFields(
				fields.NewField("error", err.Error()),
				fields.NewField("id", agentID)))
		return requestList, err
	}

	// Get retry limit and delay
	retryLimit := d.conf.SC.Get(global.ConfigRequestRetries).Int()
	retryDelay := time.Duration(d.conf.SC.Get(global.ConfigRequestRetryDelay).Int())

	// Iterate through the requests and select the ones to send
	for _, request := range requests {

		// Assume not wanted
		selected := false

		// Check if the request is new
		if request.Status == schema.RequestStatusNew {
			selected = true
		}

		// Check if the request is pending, hasn't been sent for at least global.RequestRetryTime minutes,
		// and hasn't failed more than global.RequestRetries times
		if request.Status == schema.RequestStatusPending {
			if request.LastUpdated.Before(time.Now().Add(-retryDelay*time.Minute)) && request.SendCount < retryLimit {
				selected = true
			}
		}

		if selected {
			// Validate the command before sending it to the agent
			err = commands.Validate(request.Request, request.Parameters)
			if err != nil {
				// Log the error and do not include the request in the list
				d.logger.Error(2703, "command validation failed",
					fields.NewFields(
						fields.NewField("error", err.Error()),
						fields.NewField("id", agentID),
						fields.NewField("agent", request.Request),
						fields.NewField("requestID", request.RequestID),
						fields.NewField("requester", request.Requester),
					))

				// Mark the request as failed
				request.Status = schema.RequestStatusInvalid
				updateErr := d.database.SetAgentRequest(request)
				if updateErr != nil {
					// Log the error but continue
					d.logger.Error(2707, "error marking agent request as invalid",
						fields.NewFields(
							fields.NewField("error", updateErr.Error()),
							fields.NewField("id", agentID),
							fields.NewField("requestID", request.RequestID),
							fields.NewField("requester", request.Requester),
						))
				}

			} else {

				// If the request involves downloading a file add a hash if the file exist in our server
				if request.Request == commands.DownloadExecute {
					dlURL, ok := request.Parameters["url"]
					if !ok {
						d.logger.Error(2705, "download_execute command missing url parameter",
							fields.NewFields(
								fields.NewField("id", agentID),
								fields.NewField("requestID", request.RequestID),
								fields.NewField("requester", request.Requester),
							))
						continue
					}

					// Obtain the filename from the URL
					var filename string
					filename, err = getFilename(dlURL)
					if err != nil {
						d.logger.Error(2705, "error parsing filename from download URL",
							fields.NewFields(
								fields.NewField("error", err.Error()),
								fields.NewField("id", agentID),
								fields.NewField("requestID", request.RequestID),
								fields.NewField("requester", request.Requester),
							))
						continue
					}

					// If the file does not exist, GetHash will return an empty
					// string. We'll let the agent follow its policy
					request.Parameters["hash"] = d.getHashOfFile(filename)
				}

				// if the request is an upgrade, sent the hash of the upgrade information file
				// If the file doesn't exist, this will add an empty string and the agent can
				// follow it's policy with respect to downloading the file
				if request.Request == commands.Upgrade {
					request.Parameters["hash"] = d.getHashOfFile(schema.DeployInfoFile)
				}

				// Add the request to the list
				requestList = append(requestList, schema.AgentRequest{
					Created:    request.TimeCreated,
					Requester:  request.Requester,
					RequestID:  request.RequestID,
					Request:    request.Request,
					Parameters: request.Parameters,
				})

				// Update the agent status
				if markSent {
					request.Status = schema.RequestStatusComplete
				} else {
					request.Status = schema.RequestStatusPending
				}
				request.SendCount++
				updateErr := d.database.SetAgentRequest(request)
				if updateErr != nil {
					// Log the error but continue
					d.logger.Error(2707, "error updating agent request",
						fields.NewFields(
							fields.NewField("error", updateErr.Error()),
							fields.NewField("id", agentID),
							fields.NewField("requestID", request.RequestID),
							fields.NewField("requester", request.Requester),
						))
				}
			}

		}

	}
	return requestList, nil
}

// AddAgentRequest adds a new agent for an agent
func (d *Data) AddAgentRequest(request schema.AgentRequest) (string, error) {

	// Start with the structure fields
	agentID := request.AgentID
	requestID := request.RequestID

	// If no Agent ID, look for one in the parameters
	if agentID == "" {
		if v, ok := request.Parameters[commands.AgentID]; ok {
			agentID = v
		} else {
			return "", fmt.Errorf("agent ID is required")
		}
	}

	// If no request ID, look for one in the parameters
	if requestID == "" {
		if v, ok := request.Parameters[commands.RequestID]; ok {
			requestID = v
			delete(request.Parameters, commands.RequestID)
		}
	}

	// Generate requestID if required
	if requestID == "" {
		requestID = d.generateRequestID()
	}

	// Check if the agent exists
	err := d.AgentExists(agentID)
	if err != nil {
		return "", err
	}

	// Create a new agent record in the DB
	newRequest := schema.NewDBAgentRequest()
	newRequest.AgentID = agentID
	newRequest.RequestID = requestID
	newRequest.Requester = request.Requester
	newRequest.Request = request.Request
	newRequest.AckRequired = request.AckRequired
	newRequest.Parameters = request.Parameters
	newRequest.Status = schema.RequestStatusNew
	newRequest.TimeCreated = time.Now()
	newRequest.SendCount = 0
	newRequest.Cancelled = false

	// Add the request to the database
	err = d.database.SetAgentRequest(newRequest)
	if err != nil {
		return "", fmt.Errorf("failed to add agent request: %w", err)
	}

	d.logger.Info(2706, "new agent request", fields.NewFields(
		fields.NewField("request", request.Request),
		fields.NewField("id", agentID),
		fields.NewField("requestID", requestID),
		fields.NewField("requester", request.Requester),
	))

	return newRequest.RequestID, nil
}

func (d *Data) generateRequestID() string {
	// This should always be a unique ID
	return "R-" + uuid.New().String()
}

// Get the filename from a download URL
func getFilename(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	parts := strings.Split(u.Path, "/")
	return parts[len(parts)-1], nil
}
