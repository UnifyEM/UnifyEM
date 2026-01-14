/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package user

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/UnifyEM/UnifyEM/cli/communications"
	"github.com/UnifyEM/UnifyEM/cli/display"
	"github.com/UnifyEM/UnifyEM/cli/login"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

// Register returns the root user command with subcommands.
func Register() *cobra.Command {
	userCmd := &cobra.Command{
		Use:     "user",
		Aliases: []string{"users"},
		Short:   "Manage users",
		Long:    "User management commands: list, add, delete",
	}

	userCmd.AddCommand(listCmd())
	userCmd.AddCommand(addCmd())
	userCmd.AddCommand(deleteCmd())

	return userCmd
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all users",
		RunE: func(cmd *cobra.Command, args []string) error {
			return userList()
		},
	}
}

// userList calls GET /api/v1/user and displays the result.
func userList() error {
	c := communications.New(login.Login())
	status, body, err := c.Get(schema.EndpointUser)
	return display.UserResp(status, body, err)
}

// addCmd returns the 'user add' command.
func addCmd() *cobra.Command {
	var user, displayName, email string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return userAdd(user, displayName, email)
		},
	}

	cmd.Flags().StringVarP(&user, "user", "u", "", "User (required)")
	cmd.Flags().StringVarP(&displayName, "display-name", "d", "", "Display name (optional)")
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email (required)")
	_ = cmd.MarkFlagRequired("user")
	_ = cmd.MarkFlagRequired("email")

	return cmd
}

// userAdd calls POST /api/v1/user to add a new user.
func userAdd(user, displayName, email string) error {
	if user == "" || email == "" {
		return errors.New("user and email are required")
	}
	req := schema.UserCreateRequest{
		User:        user,
		DisplayName: displayName,
		Email:       email,
	}
	c := communications.New(login.Login())
	display.ErrorWrapper(display.GenericResp(c.Post(schema.EndpointUser, req)))
	return nil
}

// deleteCmd returns the 'user delete' command.
func deleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <user_id>",
		Short: "Delete a user by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return userDelete(args[0])
		},
	}
}

// userDelete calls DELETE /api/v1/user/{id} to delete a user.
func userDelete(userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}
	c := communications.New(login.Login())
	display.ErrorWrapper(display.GenericResp(c.Delete(schema.EndpointUser + "/" + userID)))
	return nil
}
