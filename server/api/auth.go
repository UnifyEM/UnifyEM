//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/userver"
	"github.com/UnifyEM/UnifyEM/server/global"
)

// AuthInfo contains information about the authenticated user
// and their role. It implements the userver.AuthDetails interface.
type AuthInfo struct {
	ID            string // authenticated user/agent or ""
	Role          int    // authenticated role or 0
	Authenticated bool   // flag set if the user is authenticated
}

func (a AuthInfo) IsAuthenticated() bool {
	return a.Authenticated
}

// NewAuthFunc returns an AuthFunc with acceptable roles set
func (a *API) NewAuthFunc(acceptableRoles []int) userver.AuthFunc {
	return func(ip, authHeader string) (bool, []byte, any) {

		authFail := AuthInfo{
			ID:            "",
			Role:          0,
			Authenticated: false}

		// Set up log fields of interest
		logFields := fields.NewFields(fields.NewField("src_ip", ip))

		// Fail if either IP or Authorization header is missing
		if ip == "" || authHeader == "" {
			a.logger.Warning(2831, "authentication failure: missing IP or Authorization header", logFields)
			return false, a.AuthFailMessage(false), authFail
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			a.logger.Warning(2832, "authentication failure: invalid Authorization header format", logFields)
			return false, a.AuthFailMessage(false), authFail
		}

		// Extract the token from the header
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate the access token
		user, role, err := a.data.ValidateToken(tokenString, schema.TokenPurposeAccess)
		if err != nil {

			// Check if the token is expired
			if errors.Is(err, jwt.ErrTokenExpired) {
				a.logger.Info(2833, fmt.Sprintf("authentication expired: %s", err.Error()), logFields)
				return false, a.AuthFailMessage(true), authFail
			}
			a.logger.Warning(2833, fmt.Sprintf("authentication failure: %s", err.Error()), logFields)
			return false, a.AuthFailMessage(false), authFail
		}

		// Add user and role to log fields
		logFields.Append(fields.NewField("id", user), fields.NewField("role", role))

		// If the user is an admin, checked the list of authorized IP addresses
		if role == schema.RoleAdmin || role == schema.RoleSuperAdmin {
			if !a.AuthorizedAdminIP(ip) {
				a.logger.Info(2834, "authentication failure: IP not authorized", logFields)
				return false, a.AuthFailMessage(false), authFail
			}
		}

		// Check if the user's role is in the list of acceptable roles
		for _, acceptableRole := range acceptableRoles {
			if role == acceptableRole {
				a.logger.Info(2835, "authentication success", logFields)
				return true, nil, AuthInfo{ID: user, Role: role, Authenticated: true}
			}
		}

		a.logger.Warning(2836, "authentication failure: role not authorized", logFields)
		return false, a.AuthFailMessage(false), authFail
	}
}

// AuthRoles accepts a list of roles and returns them as an []int
func (a *API) AuthRoles(roles ...int) []int {
	return roles
}

// AuthAdmins is a helper function that returns a list of admin roles
func (a *API) AuthAdmins() []int {
	return []int{schema.RoleAdmin, schema.RoleSuperAdmin}
}

// AuthAnyRole returns a list of all roles
func (a *API) AuthAnyRole() []int {
	return schema.RolesAll
}

// AuthFailMessage returns a generic response for authentication failures
// The only variation is for expired tokens
func (a *API) AuthFailMessage(expired bool) []byte {

	// Start with a standard auth failure response
	msg := authFailResponse

	// If expired, update the response
	if expired {
		msg.Details = "token expired"
		msg.Status = schema.APIStatusExpired
	}

	// Marshal the response
	response, err := json.Marshal(msg)
	if err != nil {
		a.logger.Error(2839, fmt.Sprintf("error marshalling failure response: %s", err.Error()), nil)
		return nil
	}
	return response
}

func GetAuthDetails(req *http.Request) AuthInfo {
	details, ok := req.Context().Value("authDetails").(AuthInfo)
	if !ok {
		return AuthInfo{Authenticated: false}
	}
	return details
}

func (a *API) AuthorizedAdminIP(ip string) bool {

	// Get the list of authorized IPs as a map for quick lookup
	authIPList := a.conf.SC.Get(global.ConfigAuthorizedAdminIPs).SplitMap()

	// An empty list means all IPs are authorized
	if len(authIPList) == 0 {
		return true
	}

	// Check if the IP is in the list
	if _, ok := authIPList[ip]; ok {
		return true
	}
	return false
}
