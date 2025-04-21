# User Management Implementation Plan

## Background

UnifyEM currently manages agents and their metadata using BoltDB buckets, with API endpoints for agent operations. To support user management, we will introduce a new BoltDB bucket for users, a new user object schema, and a set of RESTful endpoints for CRUD operations on users. This will allow administrators to list all users, view a specific user, add a user, and delete a user.

---

## 1. Schema Design

- **User Object**: Define a new struct in `common/schema/users.go` (or similar) for user records. Example fields:
  - `UserID` (string, unique)
  - `Username` (string, unique)
  - `DisplayName` (string)
  - `Email` (string)
  - `Role` (string, e.g., admin, user)
  - `CreatedAt` (timestamp)
  - (Optional) `Tags`, `Status`, etc.

- **User List**: Define a struct for returning a list of users.

- **API Request/Response Types**: Define request/response types for user creation, deletion, and info retrieval in `common/schema`.

---

## 2. Database Layer

- **Bucket Name**: Add a new constant, e.g., `const BucketUserMeta = "UserMeta"` in the DB layer.
- **CRUD Functions**: Implement functions in the DB/data layer for:
  - Adding a user (with uniqueness check on username/UserID)
  - Retrieving a user by ID or username
  - Listing all users
  - Deleting a user

---

## 3. API Endpoints

Implement the following endpoints in `server/api/user.go`:

- `GET /user`  
  List all users. Returns a list of user objects.

- `GET /user/{id}`  
  Get info for a specific user by UserID.

- `POST /user`  
  Add a new user. Accepts a user creation request body.

- `DELETE /user/{id}`  
  Delete a user by UserID.

**Swaggo Documentation**: Add appropriate Swaggo comments for API documentation.

---

## 4. API Routing

- Register the new endpoints in the API router (e.g., in `server/api/api.go`), following the pattern used for agents.

---

## 5. CLI Integration

- Add new CLI commands in `cli/functions/user/` (e.g., `user.go`):
  - `uem-cli user list` — List all users
  - `uem-cli user get <user_id>` — Get info for a user
  - `uem-cli user add <username> <email> [--role=admin|user] [--display-name=...]` — Add a user
  - `uem-cli user delete <user_id>` — Delete a user

- Use the shared schema types for request/response.

---

## 6. Security & Validation

- Ensure only administrators can access user management endpoints.
- Validate uniqueness of usernames and required fields.
- (Optional) Add password management or authentication integration if needed.

---

## 7. Example User Object

```go
type UserMeta struct {
    UserID      string    `json:"user_id"`
    Username    string    `json:"username"`
    DisplayName string    `json:"display_name"`
    Email       string    `json:"email"`
    Role        string    `json:"role"`
    CreatedAt   time.Time `json:"created_at"`
}
```

---

## 8. Example API Usage

- **List users:**  
  `GET /user` → `[ {user1}, {user2}, ... ]`

- **Get user:**  
  `GET /user/123` → `{user1}`

- **Add user:**  
  `POST /user` with body `{ "username": "alice", "email": "alice@example.com", ... }`

- **Delete user:**  
  `DELETE /user/123`

---

## 9. Migration & Backward Compatibility

- No impact on existing agent or tag functionality.
- New bucket and endpoints are additive.

---

## 10. Next Steps

1. Define user schema and API types in `common/schema`.
2. Implement DB/data layer CRUD for users.
3. Add API handlers and register endpoints.
4. Implement CLI commands for user management.
5. Test all endpoints and CLI commands.

---

**This plan is saved as `user-implementation.md` in the docs directory for reference and recovery.**
