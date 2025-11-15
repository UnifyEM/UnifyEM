/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package data

import (
	"math/rand"
	"time"

	"github.com/UnifyEM/UnifyEM/common/schema"
)

// Auth validates a user id and password and returns the role of the user
func (d *Data) Auth(id string, pass string) (int, error) {
	role, err := d.database.CheckAuth(id, pass)
	if err != nil {
		// Impose a random delay to prevent timing attacks and make
		// brute force attacks take longer
		randomDelay()
		return schema.RoleNone, err
	}

	return role, nil
}

// SetAuth sets the password and role of a user
func (d *Data) SetAuth(id string, pass string, role int) error {
	return d.database.SetAuth(id, pass, role)
}

// LoginGetToken authenticates a user and returns access and refresh tokens or an error
func (d *Data) LoginGetToken(user string, pass string) (string, string, error) {
	var err error
	var role int
	var refreshToken, accessToken string

	// Authenticate the user
	role, err = d.Auth(user, pass)
	if err != nil {
		return "", "", err
	}

	accessToken, err = d.createToken(tokenRequest{
		subject: user,
		role:    role,
		purpose: schema.TokenPurposeAccess,
	})
	if err != nil {
		return "", "", err
	}

	refreshToken, err = d.createToken(tokenRequest{
		subject: user,
		role:    role,
		purpose: schema.TokenPurposeRefresh})
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// randomDelay imposes a random delay between 0 and 1000ms
func randomDelay() {
	delay := rand.Intn(1000)
	time.Sleep(time.Duration(delay) * time.Millisecond)
}
