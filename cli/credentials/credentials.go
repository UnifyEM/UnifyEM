/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

// Package credentials manages the access and refresh tokens. At some point in the future, they may be persisted.
package credentials

var (
	accessToken  string
	refreshToken string
)

func SetAccessToken(token string) {
	accessToken = token
}

func SetRefreshToken(token string) {
	refreshToken = token
}

func GetAccessToken() string {
	return accessToken
}

func GetRefreshToken() string {
	return refreshToken
}

func AccessExpired() {
	accessToken = ""
}

func RefreshExpired() {
	refreshToken = ""
}
