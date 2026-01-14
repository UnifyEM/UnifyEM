/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/userver"
	"github.com/UnifyEM/UnifyEM/server/reports"
)

// @Summary Generate report
// @Description Generates a report on the system
// @Tags Reporting
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body schema.ReportRequest true "Report request"
// @Success 200 {object} schema.APIReportResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 404 {object} schema.API404
// @Router /report [post]
// getReport returns the requested report
func (a *API) postReport(req *http.Request) userver.JResponse {

	remoteIP := userver.RemoteIP(req)

	authDetails := GetAuthDetails(req)

	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role))

	// Get the JSON post data
	body, err := io.ReadAll(req.Body)
	if err != nil {
		a.logger.Error(2842, fmt.Sprintf("failed reading body: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "error reading body", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Deserialize the JSON
	var cmd schema.ReportRequest
	err = json.Unmarshal(body, &cmd)
	if err != nil {
		a.logger.Error(2843, fmt.Sprintf("deserialization failed: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "error unmarshalling JSON", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Information to be logged as fields
	logFields.Append(
		fields.NewField("report", cmd.Report),
		fields.NewField("parameters", cmd.Parameters))

	// Check for missing required fields
	if cmd.Report == "" {
		a.logger.Error(2844, "missing report identifier", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "missing required fields", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Validate the command
	report, err := reports.Get(a.data, cmd)
	if err != nil {
		a.logger.Error(2845, fmt.Sprintf("report request failed: %s", err.Error()), logFields)

		details := "report request failed"

		if strings.Contains(err.Error(), "invalid report") {
			details = "requested report does not exist"
		}

		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: details, Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Log the request
	a.logger.Info(2846, "report sent", logFields)

	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APIReportResponse{
			Status:  schema.APIStatusOK,
			Code:    http.StatusOK,
			Details: "report attached",
			Report:  report}}
}
