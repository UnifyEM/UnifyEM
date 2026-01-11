/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

// Decrypt decrypts data that was encrypted with Encrypt function
// using the recipient's private encryption key
func Decrypt(encrypted string, privateEncKey string) ([]byte, error) {
	// Decode encrypted payload
	payload, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted payload: %w", err)
	}

	if len(payload) < 4 {
		return nil, fmt.Errorf("encrypted payload too short")
	}

	// Decode private key
	privKeyBytes, err := base64.StdEncoding.DecodeString(privateEncKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	privKey, err := x509.ParseECPrivateKey(privKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Extract ephemeral public key length
	ephemeralPubLen := uint16(payload[0])<<8 | uint16(payload[1])
	if len(payload) < int(4+ephemeralPubLen) {
		return nil, fmt.Errorf("invalid payload format")
	}

	// Extract ephemeral public key
	ephemeralPubBytes := payload[2 : 2+ephemeralPubLen]
	ephemeralPubInterface, err := x509.ParsePKIXPublicKey(ephemeralPubBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ephemeral public key: %w", err)
	}

	ephemeralPub, ok := ephemeralPubInterface.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("ephemeral key is not ECDSA")
	}

	// Extract encrypted AES key length
	offset := 2 + ephemeralPubLen
	encryptedKeyLen := uint16(payload[offset])<<8 | uint16(payload[offset+1])
	offset += 2

	if len(payload) < int(offset+encryptedKeyLen) {
		return nil, fmt.Errorf("invalid payload format")
	}

	// Extract encrypted AES key
	encryptedAESKey := payload[offset : offset+encryptedKeyLen]
	offset += encryptedKeyLen

	// Extract encrypted data
	encryptedData := payload[offset:]

	// Perform ECDH to get shared secret
	sharedX, _ := privKey.Curve.ScalarMult(ephemeralPub.X, ephemeralPub.Y, privKey.D.Bytes())

	// Derive decryption key from shared secret using SHA-256
	kdf := sha256.Sum256(sharedX.Bytes())

	// Decrypt AES key
	keyBlock, err := aes.NewCipher(kdf[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create key cipher: %w", err)
	}

	keyGcm, err := cipher.NewGCM(keyBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to create key GCM: %w", err)
	}

	keyNonceSize := keyGcm.NonceSize()
	if len(encryptedAESKey) < keyNonceSize {
		return nil, fmt.Errorf("encrypted AES key too short")
	}

	keyNonce := encryptedAESKey[:keyNonceSize]
	encryptedKey := encryptedAESKey[keyNonceSize:]

	aesKey, err := keyGcm.Open(nil, keyNonce, encryptedKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt AES key: %w", err)
	}

	// Decrypt data with AES-GCM
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	nonce := encryptedData[:nonceSize]
	ciphertext := encryptedData[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}
