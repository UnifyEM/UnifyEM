/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/userver"
)

// @Summary Retrieve events
// @Description Retrieves server events
// @Tags Events
// @Security BearerAuth
// @Produce json
// @Param start query string false "Start date in YYYYMMDD format"
// @Param end query string false "End date in YYYYMMDD format"
// @Param start_time query string false "Start time in Unix timestamp format"
// @Param end_time query string false "End time in Unix timestamp format"
// @Param agent_id query string false "Agent ID"
// @Success 200 {array} schema.APIEventsResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 500 {object} schema.API500
// @Router /events [get]
func (a *API) getEvents(req *http.Request) userver.JResponse {
	var err error

	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)

	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role))

	// Parse the query parameters
	query := req.URL.Query()
	agentID := query.Get("agent_id")
	start := query.Get("start")
	end := query.Get("end")
	startTimeStr := query.Get("start_time")
	endTimeStr := query.Get("end_time")
	eventType := query.Get("type")

	if agentID == "" {
		logFields.Append(fields.NewField("agent_id", agentID))
	}

	if eventType == "" {
		logFields.Append(fields.NewField("type", eventType))
	}

	if start != "" {
		logFields.Append(fields.NewField("start", start))
	}

	if end != "" {
		logFields.Append(fields.NewField("end", end))
	}

	if startTimeStr != "" {
		logFields.Append(fields.NewField("start_time", startTimeStr))
	}

	if endTimeStr != "" {
		logFields.Append(fields.NewField("end_time", endTimeStr))
	}

	// Validate the agent ID
	err = a.data.AgentExists(agentID)
	if err != nil {
		var msg string
		code := http.StatusInternalServerError
		if strings.Contains(err.Error(), "key not found") {
			msg = "agent not found"
			code = http.StatusNotFound
		} else {
			msg = fmt.Sprintf("error validating agent ID: %s", err.Error())
		}
		a.logger.Warning(2870, msg, logFields)
		return userver.JResponse{
			HTTPCode: code,
			JSONData: schema.API404{Details: msg, Status: schema.APIStatusError, Code: code}}
	}

	// Parse the start and end times
	var startT, endT, tmpInt int64

	// First check for start in YYYYMMDD format
	if start != "" {
		startTime, err := time.Parse("20060102", start)
		if err != nil {
			msg := fmt.Sprintf("invalid start date: %s", err.Error())
			logFields.Append(fields.NewField("error", msg))
			a.logger.Info(2871, "event API error", logFields)
			return userver.JResponse{
				HTTPCode: http.StatusBadRequest,
				JSONData: schema.API400{Details: msg, Status: schema.APIStatusError, Code: http.StatusBadRequest}}
		}
		startT = startTime.Unix()
	}

	// This is more precise a can therefore override the start date
	if startTimeStr != "" {
		tmpInt, err = strconv.ParseInt(startTimeStr, 10, 64)
		if err != nil {
			msg := fmt.Sprintf("invalid start_time: %s", err.Error())
			logFields.Append(fields.NewField("error", msg))
			a.logger.Info(2871, "event API error", logFields)
			return userver.JResponse{
				HTTPCode: http.StatusBadRequest,
				JSONData: schema.API400{Details: msg, Status: schema.APIStatusError, Code: http.StatusBadRequest}}
		}
		startT = tmpInt
	}

	// First check for end in YYYYMMDD format
	if end != "" {
		endTime, err := time.Parse("20060102", end)
		if err != nil {
			msg := fmt.Sprintf("invalid end date: %s", err.Error())
			logFields.Append(fields.NewField("error", msg))
			a.logger.Info(2872, "event API error", logFields)
			return userver.JResponse{
				HTTPCode: http.StatusBadRequest,
				JSONData: schema.API400{Details: msg, Status: schema.APIStatusError, Code: http.StatusBadRequest}}
		}
		endT = endTime.Unix()
	}

	if endTimeStr != "" {
		tmpInt, err = strconv.ParseInt(endTimeStr, 10, 64)
		if err != nil {
			msg := fmt.Sprintf("invalid end_time: %s", err.Error())
			logFields.Append(fields.NewField("error", msg))
			a.logger.Info(2872, "event API error", logFields)
			return userver.JResponse{
				HTTPCode: http.StatusBadRequest,
				JSONData: schema.API400{Details: msg, Status: schema.APIStatusError, Code: http.StatusBadRequest}}
		}
		endT = tmpInt
	}

	// Retrieve events based on the optional time range
	events, err := a.data.GetEvents(agentID, startT, endT, eventType)
	if err != nil {
		a.logger.Error(2873, fmt.Sprintf("error retrieving events: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusInternalServerError,
			JSONData: schema.API500{Details: "error retrieving events", Status: schema.APIStatusError, Code: http.StatusInternalServerError}}
	}

	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APIEventsResponse{
			Status: schema.APIStatusOK,
			Code:   http.StatusOK,
			Data:   events}}
}
