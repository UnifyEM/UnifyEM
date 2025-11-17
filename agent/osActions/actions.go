/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package osActions

import (
	"github.com/UnifyEM/UnifyEM/common/interfaces"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

type Actions struct {
	logger interfaces.Logger
}

func New(logger interfaces.Logger) *Actions {
	return &Actions{logger: logger}
}

func (a *Actions) Shutdown() error {
	return a.shutdownOrReboot(false)
}

func (a *Actions) Reboot() error {
	return a.shutdownOrReboot(true)
}

func (a *Actions) GetUsers() (schema.DeviceUserList, error) {
	return a.getUsers()
}

func (a *Actions) AddUser(user, password string, admin bool) error {
	return a.addUser(user, password, admin)
}

func (a *Actions) DeleteUser(user string) error {
	return a.deleteUser(user)
}

// LockUser locks out the specified user (or the current user if the
// user string is empty and optionally executes a shutdown
func (a *Actions) LockUser(user string, shutdown bool) error {

	// If shutdown option is selected, only do so if
	// locking the user account succeeds
	if shutdown {
		err := a.lockUser(user)
		if err != nil {
			return err
		}
		// Shutdown the system
		return a.shutdownOrReboot(false)
	}

	// Otherwise just lock the user's account
	return a.lockUser(user)
}

func (a *Actions) UnLockUser(user string) error {
	return a.unlockUser(user)
}

func (a *Actions) SetPassword(user, password string) error {
	return a.setPassword(user, password)
}

func (a *Actions) SetAdmin(user string, admin bool) error {
	return a.setAdmin(user, admin)
}
