/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/scrypt"
)

const (
	pemTypeEC          = "EC PRIVATE KEY"
	pemTypeEncryptedEC = "ENCRYPTED EC PRIVATE KEY"
	scryptN            = 32768
	scryptR            = 8
	scryptP            = 1
	scryptKeyLen       = 32
	saltLen            = 32
)

// GenerateSingleKeyPair generates a single P-384 keypair and returns the private and public keys
// as base64-encoded strings (same encoding as the rest of the crypto package)
func GenerateSingleKeyPair() (privateKey, publicKey string, err error) {
	_, _, privateKey, publicKey, err = GenerateKeyPairs()
	// GenerateKeyPairs returns sig and enc pairs; we reuse the enc pair for recovery
	return privateKey, publicKey, err
}

// SavePrivateKeyPEM saves a base64-encoded EC private key to a PEM file.
// If passphrase is non-empty the key is encrypted with scrypt + AES-GCM.
func SavePrivateKeyPEM(privateKeyBase64 string, path string, passphrase string) error {
	// Decode the base64 key
	keyBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return fmt.Errorf("failed to decode private key: %w", err)
	}

	var block *pem.Block

	if passphrase == "" {
		block = &pem.Block{
			Type:  pemTypeEC,
			Bytes: keyBytes,
		}
	} else {
		// Generate random salt
		salt := make([]byte, saltLen)
		if _, err := io.ReadFull(rand.Reader, salt); err != nil {
			return fmt.Errorf("failed to generate salt: %w", err)
		}

		// Derive key from passphrase using scrypt
		derivedKey, err := scrypt.Key([]byte(passphrase), salt, scryptN, scryptR, scryptP, scryptKeyLen)
		if err != nil {
			return fmt.Errorf("failed to derive key: %w", err)
		}

		// Encrypt with AES-GCM
		aesCipher, err := aes.NewCipher(derivedKey)
		if err != nil {
			return fmt.Errorf("failed to create AES cipher: %w", err)
		}
		gcm, err := cipher.NewGCM(aesCipher)
		if err != nil {
			return fmt.Errorf("failed to create GCM: %w", err)
		}
		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return fmt.Errorf("failed to generate nonce: %w", err)
		}
		ciphertext := gcm.Seal(nonce, nonce, keyBytes, nil)

		// Combine salt + ciphertext
		payload := append(salt, ciphertext...)

		block = &pem.Block{
			Type:  pemTypeEncryptedEC,
			Bytes: payload,
		}
	}

	// Write PEM file
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	return pem.Encode(f, block)
}

// LoadPrivateKeyPEM loads a PEM-encoded EC private key from a file.
// If the key is encrypted, passphrase must be provided.
// Returns the private key as a base64-encoded string.
func LoadPrivateKeyPEM(path string, passphrase string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	var keyBytes []byte

	switch block.Type {
	case pemTypeEC:
		keyBytes = block.Bytes

	case pemTypeEncryptedEC:
		if passphrase == "" {
			return "", fmt.Errorf("key is encrypted but no passphrase provided")
		}
		if len(block.Bytes) < saltLen {
			return "", fmt.Errorf("invalid encrypted key data")
		}

		salt := block.Bytes[:saltLen]
		ciphertext := block.Bytes[saltLen:]

		derivedKey, err := scrypt.Key([]byte(passphrase), salt, scryptN, scryptR, scryptP, scryptKeyLen)
		if err != nil {
			return "", fmt.Errorf("failed to derive key: %w", err)
		}

		aesCipher, err := aes.NewCipher(derivedKey)
		if err != nil {
			return "", fmt.Errorf("failed to create AES cipher: %w", err)
		}
		gcm, err := cipher.NewGCM(aesCipher)
		if err != nil {
			return "", fmt.Errorf("failed to create GCM: %w", err)
		}
		if len(ciphertext) < gcm.NonceSize() {
			return "", fmt.Errorf("ciphertext too short")
		}
		nonce := ciphertext[:gcm.NonceSize()]
		ciphertext = ciphertext[gcm.NonceSize():]

		keyBytes, err = gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt key (wrong passphrase?): %w", err)
		}

	default:
		return "", fmt.Errorf("unexpected PEM type: %s", block.Type)
	}

	// Validate it's a valid EC private key
	_, err = x509.ParseECPrivateKey(keyBytes)
	if err != nil {
		return "", fmt.Errorf("invalid EC private key: %w", err)
	}

	return base64.StdEncoding.EncodeToString(keyBytes), nil
}
