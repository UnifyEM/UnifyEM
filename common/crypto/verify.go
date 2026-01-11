/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package crypto

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"math/big"
)

// Verify verifies an ECDSA signature using SHA-256
// Returns true if signature is valid, false otherwise
//
//goland:noinspection GoUnusedExportedFunction
func Verify(data []byte, signature string, publicSignKey string) (bool, error) {
	// Decode public key
	pubKeyBytes, err := base64.StdEncoding.DecodeString(publicSignKey)
	if err != nil {
		return false, fmt.Errorf("failed to decode public key: %w", err)
	}

	pubKeyInterface, err := x509.ParsePKIXPublicKey(pubKeyBytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse public key: %w", err)
	}

	pubKey, ok := pubKeyInterface.(*ecdsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("public key is not ECDSA")
	}

	// Decode signature
	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %w", err)
	}

	if len(sigBytes) != 96 { // 48 bytes for r + 48 bytes for s
		return false, fmt.Errorf("invalid signature length")
	}

	// Extract r and s
	r := new(big.Int).SetBytes(sigBytes[:48])
	s := new(big.Int).SetBytes(sigBytes[48:])

	// Hash the data
	hash := sha256.Sum256(data)

	// Verify the signature
	valid := ecdsa.Verify(pubKey, hash[:], r, s)

	return valid, nil
}
