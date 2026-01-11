/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package db

import (
	"errors"
	"fmt"
	"time"
)

type AuthInfo struct {
	Active     bool      `json:"active"`
	HashedPass string    `json:"hashed_pass"`
	Role       int       `json:"role"`
	FailCount  int       `json:"fail_count"`
	LastUpdate time.Time `json:"time_added"`
	LastAuth   time.Time `json:"last_auth"`
	LastFail   time.Time `json:"last_fail"`
}

func NewAuthInfo() AuthInfo {
	return AuthInfo{}
}

// SetAuth stores authentication information in BucketAuth
func (d *DB) SetAuth(id string, pass string, role int) error {

	// Hash the password using bcrypt
	hashedPass, err := GenerateHash(pass)
	if err != nil {
		return fmt.Errorf("hash error: %w", err)
	}

	// Create an object
	info := NewAuthInfo()
	info.Active = true
	info.HashedPass = hashedPass
	info.Role = role
	info.LastUpdate = time.Now()

	// Use the SetData function to serialize and store the AuthInfo
	err = d.SetData(BucketAuth, validateKey(id), info)
	if err != nil {
		return fmt.Errorf("failed to store auth info: %w", err)
	}

	return nil
}

// GetAuth retrieves authentication information for a given userid
func (d *DB) GetAuth(id string) (AuthInfo, error) {
	result := NewAuthInfo()
	err := d.GetData(BucketAuth, validateKey(id), &result)
	// check if error is "key not found" and return a more useful error
	if err != nil {
		if errors.Is(err, errors.New("key not found")) {
			return result, errors.New("user not found")
		}
		return result, err
	}
	return result, err
}

// CheckAuth verifies the provided password by comparing it to the stored hashed token
// It also updates LastAuth and FailCount depending on success or failure
func (d *DB) CheckAuth(id, pass string) (int, error) {

	info, err := d.GetAuth(id)
	if err != nil {
		return 0, err
	}

	if !info.Active {
		return 0, errors.New("account disabled")
	}

	// Compare the provided password with the stored hashed token
	auth, err := VerifyHash(pass, info.HashedPass)
	if err != nil {
		return 0, fmt.Errorf("VerifyHash error: %w", err)
	}

	if auth {
		// Authentication successful
		info.FailCount = 0
		info.LastAuth = time.Now()

		// Update the record in the database
		// If this fails something is wrong - fail authorization
		if err = d.SetData(BucketAuth, validateKey(id), info); err != nil {
			return 0, err
		}

		return info.Role, nil
	}

	// Authentication failed: increment FailCount
	info.FailCount++
	info.LastFail = time.Now()

	// Update the record in the database
	if err = d.SetData(BucketAuth, validateKey(id), info); err != nil {
		return 0, err
	}

	// Return invalid password error
	return 0, errors.New("invalid password")
}

// DeleteAuth removes the authentication data for a given agent ID
func (d *DB) DeleteAuth(id string) error {
	return d.DeleteData(BucketAuth, validateKey(id))
}

// UserActive checks if a user is active
// Errors are not returned because this function is primarily used to verify
// that an agent exists and is marked active in the databaser
func (d *DB) UserActive(id string) bool {
	info, err := d.GetAuth(id)
	if err != nil {
		return false
	}

	return info.Active
}
