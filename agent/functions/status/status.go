/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

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

// UserDataSource is an interface for getting console user data from the user-helper
type UserDataSource interface {
	GetConsoleUserData() (UserContextData, bool)
}

// UserContextData represents user-specific context information
type UserContextData struct {
	Username        string
	Timestamp       time.Time
	ScreenLock      string
	ScreenLockDelay string
	RawData         map[string]string
}

type Handler struct {
	config         *global.AgentConfig
	logger         interfaces.Logger
	comms          *communications.Communications
	userDataSource UserDataSource
}

func New(config *global.AgentConfig, logger interfaces.Logger, comms *communications.Communications, userDataSource UserDataSource) *Handler {
	return &Handler{
		config:         config,
		logger:         logger,
		comms:          comms,
		userDataSource: userDataSource,
	}
}

func (h *Handler) Cmd(request schema.AgentRequest) (schema.AgentResponse, error) {
	responseData := h.CollectStatusData()
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
	f.AppendMapString(responseData.Details)

	// Log the response using separate fields
	h.logger.Info(2703, "status data", f)
	return response, nil
}

// CollectStatusData gathers all status items into AgentStatusData for reporting or testing.
func (h *Handler) CollectStatusData() schema.AgentStatusData {
	details := make(map[string]string)
	details["uem_agent"] = fmt.Sprintf("%s-%d", global.Version, global.Build)
	details["collected"] = time.Now().Format("2006-01-02T15:04:05-07:00")
	details["os"] = h.osName()
	details["os_version"] = h.osVersion()
	details["firewall"] = h.firewall()
	details["antivirus"] = h.antivirus()
	details["auto_updates"] = h.autoUpdates()
	details["full_disk_encryption"] = h.fde()
	details["password"] = h.password()
	lock, err := h.screenLock()
	if err != nil && h.logger != nil {
		h.logger.Error(2704, err.Error(), nil)
		details["screen_lock"] = "unknown"
	} else {
		details["screen_lock"] = lock
	}
	details["screen_lock_delay"] = h.screenLockDelay()
	details["hostname"] = h.hostname()
	details["last_user"] = h.lastUser()
	details["boot_time"] = h.bootTime()
	details["ip"] = h.ip()

	if global.HaveServiceAccount {
		details["service_account"] = h.checkServiceAccount()
	} else {
		details["service_account"] = "n/a"
	}

	return schema.AgentStatusData{
		Details: details,
		Info:    h.info(),
	}
}

// trapError is a helper function to log errors
func (h *Handler) trapError(value string, e error) string {
	if e != nil {
		h.logger.Error(2704, e.Error(), nil)
		return "unknown"
	}
	return value
}
