/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/hasher"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/userver"
	"github.com/UnifyEM/UnifyEM/server/global"
)

// @Summary Generate deploy.json
// @Description Creates deploy.json containing names and hashes of uem-* files
// @Tags Files
// @Security BearerAuth
// @Produce json
// @Success 200 {object} schema.APIGenericResponse
// @Failure 401 {object} schema.API401
// @Failure 500 {object} schema.API500
// @Router /files/list [post]
// @Router /files/list [put]
func (a *API) createDeployFile(req *http.Request) userver.JResponse {
	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role))

	// Get the files directory from config
	filesPath := a.conf.SC.Get(global.ConfigFilesPath).String()
	if filesPath == "" {
		a.logger.Error(2901, "files path not configured", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusInternalServerError,
			JSONData: schema.API500{
				Details: "files path not configured",
				Status:  schema.APIStatusError,
				Code:    http.StatusInternalServerError}}
	}

	// Create a map to store filename->hash pairs
	fileHashes := make(map[string]string)

	// List all files in the directory
	files, err := os.ReadDir(filesPath)
	if err != nil {
		a.logger.Error(2902, fmt.Sprintf("error reading directory: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusInternalServerError,
			JSONData: schema.API500{
				Details: "error reading directory",
				Status:  schema.APIStatusError,
				Code:    http.StatusInternalServerError}}
	}

	// Instantiate a hasher instance
	h := hasher.New()

	// Process each file
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Only process files starting with "uem-"
		if !strings.HasPrefix(file.Name(), "uem-") {
			continue
		}

		// Get the hash using existing function
		hash := h.SHA256File(filepath.Join(filesPath, file.Name())).Base64()
		if hash != "" {
			fileHashes[file.Name()] = hash
		}
	}

	// Create the deploy.json file
	deployFile := filepath.Join(filesPath, schema.DeployInfoFile)
	f, err := os.Create(deployFile)
	if err != nil {
		a.logger.Error(2903, fmt.Sprintf("error creating %s: %s", deployFile, err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusInternalServerError,
			JSONData: schema.API500{
				Details: "error creating deploy file",
				Status:  schema.APIStatusError,
				Code:    http.StatusInternalServerError}}
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	// Write the JSON data
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(fileHashes); err != nil {
		a.logger.Error(2904, fmt.Sprintf("error writing %s: %s", deployFile, err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusInternalServerError,
			JSONData: schema.API500{
				Details: "error writing deploy file",
				Status:  schema.APIStatusError,
				Code:    http.StatusInternalServerError}}
	}

	logFields.Append(fields.NewField("file_created", deployFile))
	msg := fmt.Sprintf("%s created successfully", schema.DeployInfoFile)
	a.logger.Info(2905, msg, logFields)
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APIGenericResponse{
			Status:  schema.APIStatusOK,
			Code:    http.StatusOK,
			Details: msg}}
}
