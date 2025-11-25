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

func (a *Actions) AddUser(userInfo UserInfo) error {

	// Check for invalid characters in usernames and passwords
	info, err := safeUserInfo(userInfo)
	if err != nil {
		return err
	}

	return a.addUser(info)
}

func (a *Actions) UserExists(username string) (bool, error) {

	// Check for invalid characters in username
	user, err := safeUsername(username)
	if err != nil {
		return false, err
	}

	return a.userExists(user)
}

func (a *Actions) DeleteUser(userInfo UserInfo) error {

	// Check for invalid characters in usernames and passwords
	info, err := safeUserInfo(userInfo)
	if err != nil {
		return err
	}

	return a.deleteUser(info)
}

// LockUser locks out the specified user (or the current user if the
// user string is empty and optionally executes a shutdown
func (a *Actions) LockUser(userInfo UserInfo, shutdown bool) error {

	// Check for invalid characters in usernames and passwords
	info, err := safeUserInfo(userInfo)
	if err != nil {
		return err
	}

	// If shutdown option is selected, only do so if
	// locking the user account succeeds
	if shutdown {
		err = a.lockUser(info)
		if err != nil {
			return err
		}

		// Shutdown the system
		return a.shutdownOrReboot(false)
	}

	// Otherwise just lock the user's account
	return a.lockUser(info)
}

func (a *Actions) UnLockUser(userInfo UserInfo) error {

	// Check for invalid characters in usernames and passwords
	info, err := safeUserInfo(userInfo)
	if err != nil {
		return err
	}

	return a.unlockUser(info)
}

func (a *Actions) SetPassword(userInfo UserInfo) error {

	// Check for invalid characters in usernames and passwords
	info, err := safeUserInfo(userInfo)
	if err != nil {
		return err
	}

	return a.setPassword(info)
}

func (a *Actions) SetAdmin(userInfo UserInfo) error {

	// Check for invalid characters in usernames and passwords
	info, err := safeUserInfo(userInfo)
	if err != nil {
		return err
	}

	return a.setAdmin(info)
}

func (a *Actions) TestCredentials(username string, password string) error {

	// Check for invalid characters in username
	user, err := safeUsername(username)
	if err != nil {
		return err
	}

	// Check for invalid characters in password
	pass, err := safePassword(password)
	if err != nil {
		return err
	}

	return a.testCredentials(user, pass)
}

// RefreshServiceAccount generates a new password for the service account
// It uses the old password to authenticate the change
// Returns the new password on success
func (a *Actions) RefreshServiceAccount(userInfo UserInfo) (string, error) {

	// Check for invalid characters in usernames and passwords
	info, err := safeUserInfo(userInfo)
	if err != nil {
		return "", err
	}

	return a.refreshServiceAccount(info)
}
