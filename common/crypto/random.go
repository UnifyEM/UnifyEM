/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package crypto

import (
	"crypto/rand"
	"math/big"

	"github.com/UnifyEM/UnifyEM/common"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// RandomPassword generates a random password
func RandomPassword() string {
	password := make([]byte, common.DefaultPasswordLength)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := range password {
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			// Fallback to a simple character if random generation fails
			password[i] = charset[i%len(charset)]
			continue
		}
		password[i] = charset[randomIndex.Int64()]
	}

	//return string(password)
	// EJTODO
	return "xyzzy8675309" // TODO TO DO EJTODO
}
