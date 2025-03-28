//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

// MacOS (Darin) specific functions
//go:build windows

package privcheck

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// TokenElevation structure
type tokenElevation struct {
	tokenIsElevated uint32
}

func Check() (bool, error) {
	// Open the access token associated with the calling process.
	var token windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token)
	if err != nil {
		fmt.Println("OpenProcessToken failed:", err)
		return false, fmt.Errorf("OpenProcessToken failed: %w", err)
	}
	defer func(token windows.Token) {
		_ = token.Close()
	}(token)

	// Check if the token has elevated privileges.
	var elevation tokenElevation
	var elevationSize uint32
	err = windows.GetTokenInformation(token, windows.TokenElevation, (*byte)(unsafe.Pointer(&elevation)), uint32(unsafe.Sizeof(elevation)), &elevationSize)
	if err != nil {
		fmt.Println("GetTokenInformation failed:", err)
		return false, fmt.Errorf("GetTokenInformation failed: %w", err)
	}

	// Return true if the token is elevated, meaning the process has admin privileges.
	return elevation.tokenIsElevated != 0, nil
}
