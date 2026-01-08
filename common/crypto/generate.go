/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

// GenerateKeyPairs generates two P-384 key pairs (one for signing, one for encryption)
// and returns all keys as base64-encoded strings
func GenerateKeyPairs() (privateSig, publicSig, privateEnc, publicEnc string, err error) {
	// Generate signature keypair
	sigKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to generate signature keypair: %w", err)
	}

	// Generate encryption keypair
	encKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to generate encryption keypair: %w", err)
	}

	// Encode signature private key
	sigPrivBytes, err := x509.MarshalECPrivateKey(sigKey)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to marshal signature private key: %w", err)
	}
	privateSig = base64.StdEncoding.EncodeToString(sigPrivBytes)

	// Encode signature public key
	sigPubBytes, err := x509.MarshalPKIXPublicKey(&sigKey.PublicKey)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to marshal signature public key: %w", err)
	}
	publicSig = base64.StdEncoding.EncodeToString(sigPubBytes)

	// Encode encryption private key
	encPrivBytes, err := x509.MarshalECPrivateKey(encKey)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to marshal encryption private key: %w", err)
	}
	privateEnc = base64.StdEncoding.EncodeToString(encPrivBytes)

	// Encode encryption public key
	encPubBytes, err := x509.MarshalPKIXPublicKey(&encKey.PublicKey)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to marshal encryption public key: %w", err)
	}
	publicEnc = base64.StdEncoding.EncodeToString(encPubBytes)

	return privateSig, publicSig, privateEnc, publicEnc, nil
}
