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
type UserInfo struct {
	Username      string // name of new user
	Password      string // password for new user
	Admin         bool   // should user be an administrator
	AdminUser     string // username of an existing admin (required on some platforms)
	AdminPassword string // password of the existing admin (required on some platforms)
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

func (a *Actions) AddUser(newUser UserInfo) error {
	return a.addUser(newUser)
}

func (a *Actions) UserExists(username string) (bool, error) {
	return a.userExists(username)
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

func (a *Actions) SetPassword(userInfo UserInfo) error {
	return a.setPassword(userInfo)
}

func (a *Actions) SetAdmin(user string, admin bool) error {
	return a.setAdmin(user, admin)
}

func (a *Actions) TestCredentials(user string, pass string) error {
	return a.testCredentials(user, pass)
}

// RefreshServiceAccount generates a new password for the service account
// It uses the old password to authenticate the change
// Returns the new password on success
func (a *Actions) RefreshServiceAccount(username, oldPassword string) (string, error) {
	return a.refreshServiceAccount(username, oldPassword)
}
