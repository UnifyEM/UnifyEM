package schema

import "time"

// UserMeta defines the structure for a user in UnifyEM.
type UserMeta struct {
	User        string    `json:"user" example:"alice"`
	DisplayName string    `json:"display_name,omitempty" example:"Alice Smith"`
	Email       string    `json:"email" example:"alice@example.com"`
	CreatedAt   time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	LastUpdated time.Time `json:"last_updated" example:"2023-01-01T00:00:00Z"`
}

// UserList is used to return a list of users.
// swagger:model UserList
type UserList struct {
	Users  []UserMeta `json:"users"`
	Status string     `json:"status" example:"ok"`
	Code   int        `json:"code" example:"200"`
}

// UserCreateRequest is used to create a new user.
type UserCreateRequest struct {
	User        string `json:"user"`
	DisplayName string `json:"display_name,omitempty"`
	Email       string `json:"email"`
}

// UserCreateResponse is returned after creating a user.
type UserCreateResponse struct {
	User   UserMeta `json:"user"`
	Status string   `json:"status"`
	Code   int      `json:"code"`
}

// UserDeleteResponse is returned after deleting a user.
type UserDeleteResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
}
