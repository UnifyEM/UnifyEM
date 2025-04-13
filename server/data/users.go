package data

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/server/db"
)

// AddUser adds a new user to the database. Returns error if username already exists.
func (d *Data) AddUser(user schema.UserCreateRequest) (schema.UserMeta, error) {
	// Check for existing username
	existing, _ := d.GetUserByUsername(user.User)
	if existing != nil {
		return schema.UserMeta{}, fmt.Errorf("user already exists")
	}

	now := time.Now()
	meta := schema.UserMeta{
		User:        user.User,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		CreatedAt:   now,
		LastUpdated: now,
	}
	err := d.database.SetData(db.BucketUserMeta, user.User, meta)
	if err != nil {
		return schema.UserMeta{}, err
	}
	return meta, nil
}

// GetUserByID retrieves a user by UserID.
func (d *Data) GetUserByID(user string) (*schema.UserMeta, error) {
	var meta schema.UserMeta
	err := d.database.GetData(db.BucketUserMeta, user, &meta)
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

// GetUserByUsername retrieves a user by username.
// Deprecated: Use GetUserByID instead.
func (d *Data) GetUserByUsername(user string) (*schema.UserMeta, error) {
	var meta schema.UserMeta
	err := d.database.GetData(db.BucketUserMeta, user, &meta)
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

// ListUsers returns all users in the database.
func (d *Data) ListUsers() ([]schema.UserMeta, error) {
	var users []schema.UserMeta
	err := d.database.ForEach(db.BucketUserMeta, func(_, value []byte) error {
		var meta schema.UserMeta
		if err := json.Unmarshal(value, &meta); err != nil {
			return err
		}
		users = append(users, meta)
		return nil
	})
	return users, err
}

// UserExists checks if a user exists in the user bucket.
func (d *Data) UserExists(user string) (bool, error) {
	_, err := d.GetUserByID(user)
	if err != nil {
		// If the error is "key not found", return false, nil
		if err.Error() == "key not found" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// DeleteUser deletes a user by UserID.
func (d *Data) DeleteUser(user string) error {
	return d.database.DeleteData(db.BucketUserMeta, user)
}
