/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package osActions

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/UnifyEM/UnifyEM/common/schema"
)

// safeUsername wraps a username in quotes and ensures that it only includes valid characters
//
//goland:noinspection GoUnusedFunction
func safeUsername(s string) (string, error) {
	var filtered strings.Builder
	for _, char := range s {
		if strings.ContainsRune(schema.ValidUsernameChars, char) {
			filtered.WriteRune(char)
		} else {
			return "", fmt.Errorf("invalid character '%c' in username", char)
		}
	}
	return fmt.Sprintf("\"%s\"", filtered.String()), nil
}

// safePassword wraps a password in quotes and ensures that it only includes valid characters
//
//goland:noinspection GoUnusedFunction
func safePassword(s string) (string, error) {
	var filtered strings.Builder
	for _, char := range s {
		if strings.ContainsRune(schema.ValidPasswordChars, char) {
			filtered.WriteRune(char)
		} else {
			return "", fmt.Errorf("invalid character '%c' in password", char)
		}
	}
	return fmt.Sprintf("\"%s\"", filtered.String()), nil
}

// stringClean removes all non-printable characters and multiple spaces from a string
//
//goland:noinspection GoUnusedFunction
func stringClean(b []byte) string {
	t := string(b)
	t = strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			if r == '\t' {
				return ' '
			}
			return r
		}
		return -1
	}, t)
	t = strings.Join(strings.Fields(t), " ")
	return strings.ToLower(t)
}
