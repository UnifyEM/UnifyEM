//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package status

import (
	"fmt"
	"time"

	"github.com/UnifyEM/UnifyEM/agent/communications"
	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

// Status collects and reports status information

type Handler struct {
	config *global.AgentConfig
	logger interfaces.Logger
	comms  *communications.Communications
}

func New(config *global.AgentConfig, logger interfaces.Logger, comms *communications.Communications) *Handler {
	return &Handler{
		config: config,
		logger: logger,
		comms:  comms,
	}
}

func (h *Handler) Cmd(request schema.AgentRequest) (schema.AgentResponse, error) {
	responseData := CollectStatusData(h.logger)
	response := schema.NewAgentResponse()
	response.Cmd = request.Request
	response.RequestID = request.RequestID
	response.Response = "collected"
	response.Success = true
	response.Data = responseData

	// Log the response data
	f := fields.NewFields(
		fields.NewField("cmd", request.Request),
		fields.NewField("requester", request.Requester),
		fields.NewField("request_id", request.RequestID),
	)
	f.AppendMapString(responseData)

	// Log the response using separate fields
	h.logger.Info(2703, "status data", f)
	return response, nil
}

// CollectStatusData gathers all status items into a map for reporting or testing.
func CollectStatusData(logger interfaces.Logger) map[string]string {
	responseData := make(map[string]string)
	responseData["uem_agent"] = fmt.Sprintf("%s-%d", global.Version, global.Build)
	responseData["collected"] = time.Now().Format("2006-01-02T15:04:05-07:00")
	responseData["os"] = osName()
	responseData["os_version"] = osVersion()
	responseData["firewall"] = firewall()
	responseData["antivirus"] = antivirus()
	responseData["auto_updates"] = autoUpdates()
	responseData["full_disk_encryption"] = fde()
	responseData["password"] = password()
	lock, err := screenLock()
	if err != nil && logger != nil {
		logger.Error(2704, err.Error(), nil)
		responseData["screen_lock"] = "unknown"
	} else {
		responseData["screen_lock"] = lock
	}
	responseData["screen_lock_delay"] = screenLockDelay()
	responseData["hostname"] = hostname()
	responseData["last_user"] = lastUser()
	responseData["boot_time"] = bootTime()
	responseData["ip"] = ip()
	return responseData
}

// trapError is a helper function to log errors
func (h *Handler) trapError(value string, e error) string {
	if e != nil {
		h.logger.Error(2704, e.Error(), nil)
		return "unknown"
	}
	return value
}
