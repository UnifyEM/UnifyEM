/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package crypto

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

// ValidatePublicKey verifies that the provided string is a valid base64-encoded
// PKIX EC public key, matching the format produced by GenerateKeyPairs.
func ValidatePublicKey(key string) error {
	pubKeyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return fmt.Errorf("failed to base64-decode public key: %w", err)
	}

	pubKeyInterface, err := x509.ParsePKIXPublicKey(pubKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	if _, ok := pubKeyInterface.(*ecdsa.PublicKey); !ok {
		return fmt.Errorf("public key is not an EC public key")
	}

	return nil
}
