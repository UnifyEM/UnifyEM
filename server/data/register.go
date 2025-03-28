//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package data

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"time"

	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/server/db"
	"github.com/UnifyEM/UnifyEM/server/global"
)

type RegistrationData struct {
	AgentID      string
	AccessToken  string
	RefreshToken string
}

// Register validates a registration token and returns an agent ID and password or a failure
func (d *Data) Register(regRequest schema.AgentRegisterRequest, remoteIP string) (RegistrationData, error) {
	var err error
	r := RegistrationData{}

	// Get the registration token from the config
	expectedToken := d.conf.SP.Get(global.ConfigRegToken).String()

	// Check for an empty token
	if expectedToken == "" {
		return r, errors.New("registration token is empty")
	}

	// Check if the registration token is correct
	if regRequest.Token != expectedToken {
		return r, errors.New("invalid registration token")
	}

	// Always generate a new agent ID. While it is tempting to reuse the agent ID,
	// that would allow attacker to impersonate an agent.
	r.AgentID, err = d.generateAgentID()
	if err != nil {
		return RegistrationData{}, err
	}

	// Generate new access and refresh tokens
	r.AccessToken, err = d.createToken(tokenRequest{
		subject: r.AgentID,
		role:    schema.RoleAgent,
		purpose: schema.TokenPurposeAccess})

	if err != nil {
		return RegistrationData{}, err
	}

	r.RefreshToken, err = d.createToken(tokenRequest{
		subject: r.AgentID,
		role:    schema.RoleAgent,
		purpose: schema.TokenPurposeRefresh})

	if err != nil {
		return RegistrationData{}, err
	}

	// Initialize the agent metadata
	meta := schema.NewAgentMeta(r.AgentID)
	meta.Active = true
	meta.FirstSeen = time.Now()
	meta.LastSeen = time.Now()
	meta.LastIP = remoteIP
	meta.Version = regRequest.Version
	meta.Build = regRequest.Build
	err = d.database.SetAgentMeta(meta)
	if err != nil {
		return RegistrationData{}, fmt.Errorf("db error initializing agent metadata: %w", err)
	}

	return r, nil
}

func (d *Data) generateAgentID() (string, error) {

	// This should, in theory, always be a unique ID
	id := "A-" + uuid.New().String()

	// Trust but verify...
	exists, err := d.database.KeyExists(db.BucketAgentMeta, id)
	if err != nil {
		return "", fmt.Errorf("unable to verify new agent ID: %w", err)
	}

	if exists {
		// Recurse and try again
		return d.generateAgentID()
	}
	return id, nil
}

// GenerateToken creates a random token and encodes it in base64
func (d *Data) generateToken() (string, error) {

	bytes, err := randomBytes(global.TokenLength)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// randomBytes creates size random bytes
func randomBytes(size int) ([]byte, error) {

	// Create a byte slice to hold the random data
	bytes := make([]byte, size)

	// Read random data into the byte slice
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return []byte{}, err
	}

	// Encode the byte slice in base64
	return bytes, nil
}
