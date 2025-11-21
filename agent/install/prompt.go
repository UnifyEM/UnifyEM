/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package install

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// promptCredentials prompts the user to enter administrator credentials
func (i *Install) promptCredentials() error {
	fmt.Println("The username and password of an existing administrator is required to install a service account.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	var username string
	var password string

	// Keep prompting for username until a non-empty value is provided
	for {
		fmt.Print("Username: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading username: %w", err)
		}
		username = strings.TrimSpace(input)

		if username != "" {
			break
		}
		fmt.Println("Username cannot be empty. Please try again.")
	}

	// Keep prompting for password until a non-empty value is provided
	for {
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("error reading password: %w", err)
		}
		fmt.Println() // Print newline after password input

		password = string(passwordBytes)
		if password != "" {
			break
		}
		fmt.Println("Password cannot be empty. Please try again.")
	}

	// Set the credentials
	i.user = username
	i.pass = password

	return nil
}
