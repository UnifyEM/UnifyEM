/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package recovery

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/UnifyEM/UnifyEM/cli/communications"
	"github.com/UnifyEM/UnifyEM/cli/display"
	"github.com/UnifyEM/UnifyEM/cli/global"
	"github.com/UnifyEM/UnifyEM/cli/login"
	"github.com/UnifyEM/UnifyEM/cli/util"
	"github.com/UnifyEM/UnifyEM/common/crypto"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

const defaultKeyFile = "recovery_key.pem"

func Register() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recovery",
		Short: "recovery key functions",
		Long:  "manage recovery keys and retrieve agent recovery information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("A subcommand is required\n")
			}
			return fmt.Errorf("Unknown subcommand: %s\n", args[0])
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "keygen [output_path]",
		Short: "generate recovery keypair",
		Long:  "generate a recovery keypair, save the private key to a file, and upload the public key to the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return recoveryKeygen(args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "get <agent_id> [key_path]",
		Short: "get agent recovery info",
		Long:  "retrieve and decrypt recovery information for the specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return recoveryGet(args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "check <agent_id>",
		Short: "check if recovery info exists for agent",
		Long:  "check whether recovery information has been received for the specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return recoveryCheck(args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "list recovery info status for all agents",
		Long:  "list all agents with their friendly name, agent ID, and recovery info status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return recoveryList(args, util.NewNVPairs(args))
		},
	})

	return cmd
}

func recoveryKeygen(args []string, _ *util.NVPairs) error {
	outputPath := defaultKeyFile
	if len(args) > 0 {
		outputPath = args[0]
	}

	// Generate keypair
	privateKey, publicKey, err := crypto.GenerateSingleKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate keypair: %w", err)
	}

	// Prompt for optional passphrase
	passphrase, err := promptPassphrase("Enter passphrase (leave empty for no encryption): ")
	if err != nil {
		return fmt.Errorf("failed to read passphrase: %w", err)
	}

	if passphrase != "" {
		confirm, err := promptPassphrase("Confirm passphrase: ")
		if err != nil {
			return fmt.Errorf("failed to read passphrase confirmation: %w", err)
		}
		if passphrase != confirm {
			return errors.New("passphrases do not match")
		}
	}

	// Save private key to PEM file
	err = crypto.SavePrivateKeyPEM(privateKey, outputPath, passphrase)
	if err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}
	fmt.Printf("Private key saved to: %s\n", outputPath)

	// Upload public key to server
	c := communications.New(login.Login())
	keyReq := schema.RecoveryKeyRequest{PublicKey: publicKey}
	display.ErrorWrapper(display.GenericResp(c.Post(schema.EndpointRecovery+"/key", keyReq)))
	return nil
}

func recoveryGet(args []string, _ *util.NVPairs) error {
	if len(args) == 0 {
		return errors.New("agent ID is required\n")
	}

	agentID := args[0]
	keyPath := defaultKeyFile
	if len(args) > 1 {
		keyPath = args[1]
	}

	// Fetch encrypted blob from server
	c := communications.New(login.Login())
	statusCode, data, err := c.Get(schema.EndpointAgent + "/" + agentID + "/recovery")
	if err != nil {
		return fmt.Errorf("failed to retrieve recovery info: %w", err)
	}

	fmt.Printf("\nServer response: HTTP %d\n", statusCode)

	var resp schema.APIRecoveryResponse
	if err = json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.RecoveryInfo == "" {
		global.Pretty(resp)
		return nil
	}

	// Try to load the private key without passphrase first
	privateKey, err := crypto.LoadPrivateKeyPEM(keyPath, "")
	if err != nil {
		// Only prompt for a passphrase if the key file exists but is encrypted;
		// any other error (e.g. file not found) is returned immediately.
		if !strings.Contains(err.Error(), "encrypted") {
			return fmt.Errorf("failed to load private key: %w", err)
		}
		passphrase, pErr := promptPassphrase("Enter passphrase for private key: ")
		if pErr != nil {
			return fmt.Errorf("failed to read passphrase: %w", pErr)
		}
		privateKey, err = crypto.LoadPrivateKeyPEM(keyPath, passphrase)
		if err != nil {
			return fmt.Errorf("failed to load private key: %w", err)
		}
	}

	// Decrypt the blob
	plaintext, err := crypto.Decrypt(resp.RecoveryInfo, privateKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt recovery info: %w", err)
	}

	// Unmarshal and display
	var info schema.RecoveryInfo
	if err = json.Unmarshal(plaintext, &info); err != nil {
		return fmt.Errorf("failed to unmarshal recovery info: %w", err)
	}

	fmt.Println("\nRecovery Information:")
	global.Pretty(info)
	return nil
}

func recoveryCheck(args []string, _ *util.NVPairs) error {
	if len(args) == 0 {
		return errors.New("agent ID is required\n")
	}

	agentID := args[0]

	c := communications.New(login.Login())
	statusCode, data, err := c.Get(schema.EndpointAgent + "/" + agentID + "/recovery")
	if err != nil {
		return fmt.Errorf("failed to retrieve recovery info: %w", err)
	}

	fmt.Printf("\nServer response: HTTP %d\n", statusCode)

	var resp schema.APIRecoveryResponse
	if err = json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.RecoveryInfo != "" {
		fmt.Printf("Recovery info available\n")
	} else {
		fmt.Printf("No recovery info available\n")
	}

	return nil
}

func recoveryList(_ []string, _ *util.NVPairs) error {
	c := communications.New(login.Login())
	statusCode, data, err := c.Get(schema.EndpointAgent)
	if err != nil {
		return fmt.Errorf("failed to retrieve agent list: %w", err)
	}

	fmt.Printf("\nServer response: HTTP %d\n", statusCode)

	var resp schema.APIAgentInfoResponse
	if err = json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(resp.Data.Agents) == 0 {
		fmt.Printf("No agents found\n")
		return nil
	}

	for _, agent := range resp.Data.Agents {
		recoveryStatus := "No"
		if agent.RecoveryInfo != "" {
			recoveryStatus = "Yes"
		}
		fmt.Printf("%-30s %-36s %s\n", agent.FriendlyName, agent.AgentID, recoveryStatus)
	}

	return nil
}

// promptPassphrase reads a line from stdin (passphrase will be visible)
func promptPassphrase(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
