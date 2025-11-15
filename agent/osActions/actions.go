/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package osActions

import (
	"errors"

	"github.com/UnifyEM/UnifyEM/agent/global"
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

// LockUser locks out the specified user (or the current user if the
// user string is empty and optionally executes a shutdown
func (a *Actions) LockUser(user string, shutdown bool) error {
	if global.PROTECTED {
		return errors.New("LockUser is disabled in protected mode")
	}

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
	if global.PROTECTED {
		return errors.New("UnLockUser is disabled in protected mode")
	}
	return a.unlockUser(user)
}

func (a *Actions) SetPassword(user, password string) error {
	if global.PROTECTED {
		return errors.New("SetPassword is disabled in protected mode")
	}
	return a.setPassword(user, password)
}

func (a *Actions) SetAdmin(user string, admin bool) error {
	if global.PROTECTED {
		return errors.New("SetAdmin is disabled in protected mode")
	}
	return a.setAdmin(user, admin)
}
