/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
)

// Encrypt encrypts data using hybrid encryption:
// 1. Generate random AES-GCM key
// 2. Encrypt data with AES-GCM
// 3. Encrypt AES key with recipient's P-384 public key using ECIES
// Returns base64-encoded encrypted payload
func Encrypt(data []byte, recipientPublicEncKey string) (string, error) {
	// Decode recipient's public key
	pubKeyBytes, err := base64.StdEncoding.DecodeString(recipientPublicEncKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode public key: %w", err)
	}

	pubKeyInterface, err := x509.ParsePKIXPublicKey(pubKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %w", err)
	}

	pubKey, ok := pubKeyInterface.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("public key is not ECDSA")
	}

	// Generate random AES-256 key
	aesKey := make([]byte, 32) // 256 bits
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return "", fmt.Errorf("failed to generate AES key: %w", err)
	}

	// Encrypt data with AES-GCM
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	encryptedData := gcm.Seal(nonce, nonce, data, nil)

	// Encrypt AES key using ECIES (simplified version using ECDH + KDF)
	// Generate ephemeral keypair
	ephemeralKey, err := ecdsa.GenerateKey(pubKey.Curve, rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate ephemeral key: %w", err)
	}

	// Perform ECDH to get shared secret
	sharedX, _ := pubKey.Curve.ScalarMult(pubKey.X, pubKey.Y, ephemeralKey.D.Bytes())

	// Derive encryption key from shared secret using SHA-256
	kdf := sha256.Sum256(sharedX.Bytes())

	// Encrypt AES key with derived key
	keyBlock, err := aes.NewCipher(kdf[:])
	if err != nil {
		return "", fmt.Errorf("failed to create key cipher: %w", err)
	}

	keyGcm, err := cipher.NewGCM(keyBlock)
	if err != nil {
		return "", fmt.Errorf("failed to create key GCM: %w", err)
	}

	keyNonce := make([]byte, keyGcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, keyNonce); err != nil {
		return "", fmt.Errorf("failed to generate key nonce: %w", err)
	}

	encryptedAESKey := keyGcm.Seal(keyNonce, keyNonce, aesKey, nil)

	// Marshal ephemeral public key
	ephemeralPubBytes, err := x509.MarshalPKIXPublicKey(&ephemeralKey.PublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ephemeral public key: %w", err)
	}

	// Combine: ephemeralPubKey length (2 bytes) + ephemeralPubKey + encryptedAESKey length (2 bytes) + encryptedAESKey + encryptedData
	ephemeralPubLen := uint16(len(ephemeralPubBytes))
	encryptedKeyLen := uint16(len(encryptedAESKey))

	payload := make([]byte, 0, 4+len(ephemeralPubBytes)+len(encryptedAESKey)+len(encryptedData))
	payload = append(payload, byte(ephemeralPubLen>>8), byte(ephemeralPubLen))
	payload = append(payload, ephemeralPubBytes...)
	payload = append(payload, byte(encryptedKeyLen>>8), byte(encryptedKeyLen))
	payload = append(payload, encryptedAESKey...)
	payload = append(payload, encryptedData...)

	return base64.StdEncoding.EncodeToString(payload), nil
}
