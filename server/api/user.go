/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
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
)

// @Summary List all users
// @Description Retrieves a list of all users
// @Tags User management
// @Security BearerAuth
// @Produce json
// @Success 200 {object} schema.UserList
// @Failure 401 {object} schema.API401
// @Failure 500 {object} schema.API500
// @Router /user [get]
func (a *API) getUsers(req *http.Request) userver.JResponse {
	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role),
	)

	users, err := a.data.ListUsers()
	if err != nil {
		a.logger.Error(3201, fmt.Sprintf("error retrieving users: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusInternalServerError,
			JSONData: schema.API500{Details: "error retrieving users", Status: "error", Code: http.StatusInternalServerError}}
	}
	a.logger.Info(3202, "users listed", logFields)
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.UserList{Users: users, Status: "ok", Code: http.StatusOK},
	}
}

// @Summary Get user by ID
// @Description Retrieves information for a specific user
// @Tags User management
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} schema.UserMeta
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 404 {object} schema.API404
// @Failure 500 {object} schema.API500
// @Router /user/{id} [get]
func (a *API) getUser(req *http.Request) userver.JResponse {
	userID := userver.GetParam(req, "id")
	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role),
		fields.NewField("user_id", userID),
	)

	if userID == "" {
		a.logger.Error(3203, "user ID required", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "user ID required", Status: "error", Code: http.StatusBadRequest}}
	}
	user, err := a.data.GetUserByID(userID)
	if err != nil {
		code := http.StatusNotFound
		status := "not found"
		if !strings.Contains(err.Error(), "not found") {
			code = http.StatusInternalServerError
			status = "error"
		}
		a.logger.Error(3204, fmt.Sprintf("error retrieving user: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: code,
			JSONData: schema.UserList{
				Users:  []schema.UserMeta{},
				Status: status,
				Code:   code,
			},
		}
	}
	a.logger.Info(3205, "user retrieved", logFields)
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.UserList{
			Users:  []schema.UserMeta{*user},
			Status: "ok",
			Code:   http.StatusOK,
		},
	}
}

// @Summary Add a new user
// @Description Adds a new user
// @Tags User management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user body schema.UserCreateRequest true "User data"
// @Success 200 {object} schema.UserCreateResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 409 {object} schema.API400
// @Failure 500 {object} schema.API500
// @Router /user [post]
func (a *API) postUser(req *http.Request) userver.JResponse {
	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role),
	)

	body, err := io.ReadAll(req.Body)
	if err != nil {
		a.logger.Error(3206, "error reading body", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "error reading body", Status: "error", Code: http.StatusBadRequest}}
	}
	var createReq schema.UserCreateRequest
	if err := json.Unmarshal(body, &createReq); err != nil {
		a.logger.Error(3207, "error unmarshalling JSON", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "error unmarshalling JSON", Status: "error", Code: http.StatusBadRequest}}
	}
	user, err := a.data.AddUser(createReq)
	if err != nil {
		code := http.StatusInternalServerError
		if strings.Contains(err.Error(), "exists") {
			code = http.StatusConflict
		}
		a.logger.Error(3208, err.Error(), logFields)
		return userver.JResponse{
			HTTPCode: code,
			JSONData: schema.API400{Details: err.Error(), Status: "error", Code: code}}
	}
	logFields.Append(fields.NewField("user", user.User))
	a.logger.Info(3209, "user added", logFields)
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.UserCreateResponse{User: user, Status: "ok", Code: http.StatusOK},
	}
}

// @Summary Delete a user
// @Description Deletes a user by ID
// @Tags User management
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} schema.UserDeleteResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 404 {object} schema.API404
// @Failure 500 {object} schema.API500
// @Router /user/{id} [delete]
func (a *API) deleteUser(req *http.Request) userver.JResponse {
	userID := userver.GetParam(req, "id")
	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role),
		fields.NewField("user_id", userID),
	)

	if userID == "" {
		a.logger.Error(3210, "user ID required", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "user ID required", Status: "error", Code: http.StatusBadRequest}}
	}
	err := a.data.DeleteUser(userID)
	if err != nil {
		code := http.StatusInternalServerError
		status := "error"
		if strings.Contains(err.Error(), "not found") {
			code = http.StatusNotFound
			status = "not found"
		}
		a.logger.Error(3211, "user not found", logFields)
		return userver.JResponse{
			HTTPCode: code,
			JSONData: schema.API404{Details: "user not found", Status: status, Code: code}}
	}
	a.logger.Info(3212, "user deleted", logFields)
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.UserDeleteResponse{Status: "ok", Code: http.StatusOK},
	}
}
