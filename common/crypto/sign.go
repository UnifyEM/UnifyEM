/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

// Sign signs data using ECDSA with SHA-256
// Returns base64-encoded signature
//
//goland:noinspection GoUnusedExportedFunction
func Sign(data []byte, privateSignKey string) (string, error) {
	// Decode private key
	privKeyBytes, err := base64.StdEncoding.DecodeString(privateSignKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %w", err)
	}

	privKey, err := x509.ParseECPrivateKey(privKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// Hash the data
	hash := sha256.Sum256(data)

	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign data: %w", err)
	}

	// Encode signature (r || s)
	rBytes := r.Bytes()
	sBytes := s.Bytes()

	// Ensure both r and s are 48 bytes (384 bits / 8)
	signature := make([]byte, 96)
	copy(signature[48-len(rBytes):48], rBytes)
	copy(signature[96-len(sBytes):96], sBytes)

	return base64.StdEncoding.EncodeToString(signature), nil
}
