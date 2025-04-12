package schema

import "time"

// UserMeta defines the structure for a user in UnifyEM.
// UserMeta defines the structure for a user in UnifyEM.
// UserMeta defines the structure for a user in UnifyEM.
type UserMeta struct {
	User        string    `json:"user"`
	DisplayName string    `json:"display_name,omitempty"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
	LastUpdated time.Time `json:"last_updated"`
}

// UserList is used to return a list of users.
// swagger:model UserList
// UserList is used to return a list of users.
// swagger:model UserList
// UserList is used to return a list of users.
// swagger:model UserList
type UserList struct {
	// The list of users
	Users []UserMeta `json:"users" example:"[{\"user\":\"alice\",\"display_name\":\"Alice\",\"email\":\"alice@example.com\",\"created_at\":\"2024-01-01T00:00:00Z\"}]"`
	// Status of the response
	Status string `json:"status" example:"ok"`
	// HTTP status code
	Code int `json:"code" example:"200"`
}

// UserCreateRequest is used to create a new user.
// UserCreateRequest is used to create a new user.
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
