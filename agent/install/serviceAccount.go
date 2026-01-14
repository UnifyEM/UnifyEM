/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package install

import (
	"fmt"

	"github.com/UnifyEM/UnifyEM/agent/osActions"
	"github.com/UnifyEM/UnifyEM/common"
	uemCrypto "github.com/UnifyEM/UnifyEM/common/crypto"
	"github.com/UnifyEM/UnifyEM/common/fields"
)

func (i *Install) ServiceAccount() error {
	var err error

	// Create an osActions instance
	actions := osActions.New(i.logger)

	// Does the service user already exist?
	var exists bool
	exists, err = actions.UserExists(common.ServiceAccount)
	if err != nil {
		return fmt.Errorf("error checking if %s exists: %w", common.ServiceAccount, err)
	}

	if exists {
		return i.refreshServiceAccount(actions)
	}
	return i.createServiceAccount(actions)
}

func (i *Install) createServiceAccount(actions *osActions.Actions) error {

	newPassword := uemCrypto.RandomPassword()

	err := actions.AddUser(
		osActions.UserInfo{
			Username:      common.ServiceAccount,
			Password:      newPassword,
			Admin:         true,
			AdminUser:     i.user,
			AdminPassword: i.pass})

	if err != nil {
		return fmt.Errorf("error creating service account %s: %w", common.ServiceAccount, err)
	}

	// Store encrypted credentials in config
	err = i.config.SetServiceCredentials(common.ServiceAccount, newPassword)
	if err != nil {
		i.logger.Warningf(8111, "failed to store service credentials: %s", err.Error())
		return fmt.Errorf("failed to store service credentials: %w", err)
	}

	i.logger.Info(8112, "service credentials encrypted and stored in memory", nil)
	i.logger.Info(8418, "service account created with random password",
		fields.NewFields(fields.NewField("account", common.ServiceAccount)))

	// Send credentials to server
	err = i.sendServiceCredentialsToServer()
	if err != nil {
		return fmt.Errorf("failed to send service credentials to server: %w", err)
	}

	return nil
}

func (i *Install) refreshServiceAccount(actions *osActions.Actions) error {
	newPassword := uemCrypto.RandomPassword()

	userInfo := osActions.UserInfo{
		Username:      common.ServiceAccount,
		Password:      newPassword,
		Admin:         true,
		AdminUser:     i.user,
		AdminPassword: i.pass}

	err := actions.SetAdmin(userInfo)
	if err != nil {
		return fmt.Errorf("error setting service account %s as admin: %w", common.ServiceAccount, err)
	}

	err = actions.SetPassword(userInfo)

	if err != nil {
		return fmt.Errorf("error updating service account %s: %w", common.ServiceAccount, err)
	}

	// Store encrypted credentials in config
	err = i.config.SetServiceCredentials(common.ServiceAccount, newPassword)
	if err != nil {
		i.logger.Warningf(8113, "failed to store service credentials: %s", err.Error())
		return fmt.Errorf("failed to store service credentials: %w", err)
	}

	i.logger.Info(8114, "service credentials encrypted and stored in memory", nil)
	i.logger.Info(8419, "service account updated with new random password",
		fields.NewFields(fields.NewField("account", common.ServiceAccount)))

	// Send credentials to server
	err = i.sendServiceCredentialsToServer()
	if err != nil {
		return fmt.Errorf("failed to send service credentials to server: %w", err)
	}

	return nil
}
