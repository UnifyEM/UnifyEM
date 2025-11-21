/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package data

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/server/global"
)

// CustomClaims includes jwt.RegisteredClaims and adds custom fields
type CustomClaims struct {
	jwt.RegisteredClaims
	Role    int    `json:"role"`
	Purpose string `json:"purpose"`
}

type tokenRequest struct {
	subject string
	role    int
	purpose string
}

// createToken requires the subject, role, and lifetime of the JWT in minutes
func (d *Data) createToken(request tokenRequest) (string, error) {
	var lifeTime int

	// Get the appropriate lifetime
	switch request.purpose {
	case schema.TokenPurposeAccess:
		lifeTime = d.conf.SC.Get(global.ConfigAccessTokenLife).Int()
	case schema.TokenPurposeRefresh:
		if request.role == schema.RoleAgent {
			lifeTime = d.conf.SC.Get(global.ConfigRefreshTokenLifeAgents).Int()
		} else {
			lifeTime = d.conf.SC.Get(global.ConfigRefreshTokenLifeUsers).Int()
		}
	default:
		return "", errors.New("invalid token purpose")
	}

	// Define the JWT claims
	// Set NotBefore 5 minutes in the past to allow for clock skew
	now := time.Now()
	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   request.subject,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Minute)),
			Issuer:    global.Name,
			ID:        "T-" + uuid.New().String(),
		},
		Role:    request.role,
		Purpose: request.purpose,
	}

	// If token lifetime is limited, add the expiration time/date
	if lifeTime > 0 {
		claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(now.Add(time.Duration(lifeTime) * time.Minute))
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString(d.jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ValidateToken validates the supplied token (including purpose) and returns the user, role, and error
func (d *Data) ValidateToken(tokenString string, purpose string) (string, int, error) {

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return d.jwtKey, nil
	})
	if err != nil {
		return "", 0, err
	}

	// Validate the token and extract the claims
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		// Check the purpose
		if claims.Purpose == purpose {
			return claims.Subject, claims.Role, nil
		}
	}
	return "", 0, errors.New("invalid token")
}

type TokenRefreshData struct {
	AccessToken     string
	ServerPublicSig string
	ServerPublicEnc string
}

// RefreshToken returns a new access token or an error
// If clientPublicSig and clientPublicEnc are provided (not empty), they will be updated in the agent metadata (for rekey scenarios)
func (d *Data) RefreshToken(refreshToken string, clientPublicSig string, clientPublicEnc string) (TokenRefreshData, error) {
	var err error
	var role int
	var accessToken string

	// Validate the refresh token
	subject, role, err := d.ValidateToken(refreshToken, schema.TokenPurposeRefresh)
	if err != nil {
		return TokenRefreshData{}, err
	}

	// Check if the agent or user exists and is marked active
	subjectActive := false
	if role == schema.RoleAgent {
		subjectActive = d.database.AgentActive(subject)
	} else {
		subjectActive = d.database.UserActive(subject)
	}

	if !subjectActive {
		return TokenRefreshData{}, fmt.Errorf("subject disabled in database: %s", subject)
	}

	// If this is an agent and new client public keys are provided, update them (for rekey scenarios)
	if role == schema.RoleAgent && (clientPublicSig != "" || clientPublicEnc != "") {
		meta, err := d.database.GetAgentMeta(subject)
		if err != nil {
			return TokenRefreshData{}, fmt.Errorf("failed to get agent metadata: %w", err)
		}

		// Update keys if provided
		if clientPublicSig != "" {
			meta.ClientPublicSig = clientPublicSig
		}
		if clientPublicEnc != "" {
			meta.ClientPublicEnc = clientPublicEnc
		}

		err = d.database.SetAgentMeta(meta)
		if err != nil {
			return TokenRefreshData{}, fmt.Errorf("failed to update agent metadata: %w", err)
		}
	}

	// Create a new access token
	accessToken, err = d.createToken(tokenRequest{
		subject: subject,
		role:    role,
		purpose: schema.TokenPurposeAccess})
	if err != nil {
		return TokenRefreshData{}, err
	}

	// Get server public keys from configuration
	serverPublicSig := d.conf.SP.Get(global.ConfigServerECPublicSig).String()
	serverPublicEnc := d.conf.SP.Get(global.ConfigServerECPublicEnc).String()

	return TokenRefreshData{
		AccessToken:     accessToken,
		ServerPublicSig: serverPublicSig,
		ServerPublicEnc: serverPublicEnc,
	}, nil
}
