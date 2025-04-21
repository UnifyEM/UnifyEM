# UnifyEM API Best Practices & Requirements

## API Endpoint Conventions

- All API endpoints are registered in `server/api/api.go` using `s.AddRoute`.

## User Field Best Practices

- The `user` field is the unique identifier for users and is used as the key in the database.
- There is no `user_id` field; do not add or use one.
- All code, API, and CLI must use the `user` field as the identifier for add, get, list, and delete operations.
- The `user` field replaces any previous use of `username` everywhere.
- All user-related operations (add, get, list, delete) must use `user` as the key and identifier.
- All examples, documentation, and log fields must use `user` (not `user_id` or `username`).

- Endpoints use constants from `common/schema/apiMeta.go` (e.g., `schema.EndpointUser`).
- Endpoints should be namespaced under `/api/v1/` (e.g., `/api/v1/user`).

## API Response Struct Requirements

- All structs returned from the API **must** include:
  - `Status` (string): Indicates the status of the response (e.g., "ok", "error", "not found", "expired").
  - `Code` (int): The HTTP status code (e.g., 200, 404).
- All fields must have appropriate `json` tags.
- Example values for all fields should be provided using `example` struct tags for Swagger/OpenAPI documentation.
- Add a `// swagger:model StructName` comment above each struct for documentation generation.
- See `common/schema/apiResponses.go` and `common/schema/users.go` for examples.
- **User records must always include both `CreatedAt` and `LastUpdated` fields. These must be set by the backend (data layer) and must never be settable or modifiable by the admin or API client.**
- **The `user` field is the unique identifier for users and is used as the key in the database. There is no `user_id` field. All code, API, and CLI must use `user` as the identifier.**

### Example

```go
// swagger:model UserMeta
type UserMeta struct {
    User        string    `json:"user"`
    DisplayName string    `json:"display_name,omitempty"`
    Email       string    `json:"email"`
    CreatedAt   time.Time `json:"created_at"`
    LastUpdated time.Time `json:"last_updated"`
}
```

## User Identifier Conventions

- The `user` field is the unique identifier for users and is used as the key in the database.
- There is no `user_id` field; do not add or use one.
- All code, API, and CLI must use the `user` field as the identifier for add, get, list, and delete operations.
- The `user` field replaces any previous use of `username`.

## API Handler Patterns

- For "get" endpoints (e.g., get user by ID), always return a list object (e.g., `UserList` with one user) for consistency.
- For "list" endpoints, return the same list object with all results.
- For simple status replies (e.g., 200 OK, 404 Not Found), use `APIGenericResponse` or similar.
- Always set the correct HTTP status code in the response and in the `Code` field of the struct.

## Error Handling

- If a resource is not found, return an empty list and set `Status` to `"not found"` and `Code` to `404`.
- For other errors, set `Status` to `"error"` and `Code` to the appropriate HTTP status.

## API Logging Conventions

- All API handlers must log both errors and successful operations.
- **Every successful create, update, or delete operation must log an Info-level entry immediately before returning the success JSON response.**
- Use a unique numeric code for each log entry (e.g., start user-related codes at 3201).
- Use the appropriate log level:
  - `a.logger.Error(code, message, logFields)` for errors
  - `a.logger.Info(code, message, logFields)` for successful operations
  - `a.logger.Warning(code, message, logFields)` for warnings
- Build `logFields` using `fields.NewFields`, including at least:
  - `src_ip` (from `userver.RemoteIP(req)`)
  - `id` and `role` (from `GetAuthDetails(req)`)
  - Any other relevant fields (e.g., `user_id` for user operations)
- Always log the reason for failure, including the error message.
- See `server/api/agent.go` and `server/api/user.go` for examples.

### Example

```go
remoteIP := userver.RemoteIP(req)
authDetails := GetAuthDetails(req)
logFields := fields.NewFields(
    fields.NewField("src_ip", remoteIP),
    fields.NewField("id", authDetails.ID),
    fields.NewField("role", authDetails.Role),
    fields.NewField("user", user),
)

if err != nil {
    a.logger.Error(3201, fmt.Sprintf("error retrieving users: %s", err.Error()), logFields)
    // ...
}
a.logger.Info(3202, "users listed", logFields)
```

## CLI Display Conventions

- CLI display functions should be placed in `cli/display/` (e.g., `userResp.go`, `tagsResp.go`).
- Display functions should:
  - Unmarshal the response into the correct schema struct.
  - Check for expired access tokens and call `credentials.AccessExpired()` if needed.
  - Use `global.Pretty()` to print the struct.
  - Print the HTTP status code.
- Example display function: see `cli/display/userResp.go`.

## Authentication

- Most endpoints require BearerAuth (JWT) and should be protected accordingly.
- The CLI uses the `login` package to manage authentication and tokens.

## Database Bucket Conventions

- Every BoltDB bucket must be defined as a `const` in `server/db/db.go` (e.g., `const BucketUsers = "Users"`).
- The bucket const must be added to the `bucketList` slice in `db.go` to ensure it is created on startup.
- All code must reference the bucket using the const, never a string literal.
- This ensures consistency, prevents typos, and guarantees all required buckets are created.
- See `server/db/db.go` for examples.

## General Notes

- Do **not** include password fields in user-related API requests or responses.
- User creation and update timestamps (`CreatedAt`, `LastUpdated`) must be set by the backend only and never exposed for modification via the API or CLI.
- Always keep API and CLI response structures in sync.
- Update Swagger/OpenAPI comments and example tags whenever you change a struct.

---

## Agents and Users: Model, Endpoints, and Assignment

### Agent and User Schema

#### UserMeta

- The user object is defined in `common/schema/users.go` as `UserMeta`:
  ```go
  // swagger:model UserMeta
  type UserMeta struct {
      User        string    `json:"user"`
      DisplayName string    `json:"display_name,omitempty"`
      Email       string    `json:"email"`
      CreatedAt   time.Time `json:"created_at"`
      LastUpdated time.Time `json:"last_updated"`
  }
  ```
- The `user` field is the unique identifier for users and is used as the key in the database.

#### AgentMeta

- The agent object is defined in `common/schema/agents.go` as `AgentMeta`:
  ```go
  // swagger:model AgentMeta
  type AgentMeta struct {
      AgentID      string        `json:"agent_id"`
      Active       bool          `json:"active"`
      FriendlyName string        `json:"friendly_name"`
      FirstSeen    time.Time     `json:"first_seen"`
      LastSeen     time.Time     `json:"last_seen"`
      LastIP       string        `json:"last_ip"`
      Version      string        `json:"version"`
      Build        int           `json:"build"`
      Triggers     AgentTriggers `json:"triggers"`
      Status       *AgentStatus  `json:"status,omitempty"`
      Tags         []string      `json:"tags"`
      Users        []string      `json:"users" example:"[\"alice\",\"bob\"]"`
  }
  ```
- The `users` field is a list of usernames (strings) assigned to this agent, representing which users have access to the computer/device.

### API Endpoints for Agents and Users

#### User Endpoints

- `GET /user` — List all users.
- `GET /user/{user}` — Get details for a specific user.
- `POST /user` — Add a new user.
- `DELETE /user/{user}` — Delete a user.

#### Agent Endpoints

- `GET /agent` — List all agents.
- `GET /agent/{id}` — Get details for a specific agent (including the `users` field).
- `POST /agent/{id}` or `PUT /agent/{id}` — Update agent information.
- `DELETE /agent/{id}` — Delete an agent.
- `GET /agent/by-tag/{tag}` — List all agents with a specific tag.

#### Agent User Assignment Endpoints

- `POST /agent/{id}/users/add` — Add one or more users to an agent.
  - Request body: `{ "users": ["alice", "bob"] }`
  - Checks that each user exists before adding.
  - Duplicates are ignored.
  - Logs every error and every success.
- `POST /agent/{id}/users/remove` — Remove one or more users from an agent.
  - Request body: `{ "users": ["alice"] }`
  - Removes the specified users from the agent's `users` list.
  - Logs every error and every success.

### CLI Commands for User-Agent Assignment

- `agent user-add <agent_id> <user1> [<user2> ...]` — Add users to a specific agent.
- `agent user-remove <agent_id> <user1> [<user2> ...]` — Remove users from a specific agent.
- `agent user-add tag=<tag> <user1> [<user2> ...]` — Add users to all agents with the specified tag.
- `agent user-remove tag=<tag> <user1> [<user2> ...]` — Remove users from all agents with the specified tag.

The CLI will:
- For `tag=<tag>`, query `/agent/by-tag/{tag}` to get all agents with that tag, then perform the add/remove operation for each agent.
- For a specific agent ID, perform the operation directly.

### Data Flow for Adding/Removing Users to/from Agents

1. The CLI command is invoked (e.g., `agent user-add agent123 alice`).
2. The CLI sends a POST request to `/agent/agent123/users/add` with the user list.
3. The API handler:
   - Checks that each user exists in the user bucket.
   - Adds users to the agent's `users` list, ensuring uniqueness.
   - Updates the agent metadata in the database.
   - Logs every error and every success.
   - Returns the updated list of users for the agent.
4. For `tag=<tag>`, the CLI first queries `/agent/by-tag/{tag}` and repeats the above process for each agent.

### Requirements and Conventions for Future Changes

- The `users` field in `AgentMeta` must always be a list of strings, with the JSON tag `users`.
- All user-related operations must use the `user` field as the unique identifier.
- When adding users to an agent, always check that the user exists before adding.
- All API handlers must log both errors and successful operations, using unique numeric codes.
- The agent's `users` list is the source of truth for which users have access to a device.
- Do not add a `user_id` field or use any identifier other than `user`.
- When making changes to user or agent assignment logic, ensure that:
  - The API, CLI, and schema remain in sync.
  - All changes are reflected in Swagger/OpenAPI documentation and example tags.
  - Logging and error handling conventions are followed.
- If you need to add additional actions when users are added/removed from agents, use the HOOK comments in the API code as insertion points.

### Example: Adding Users to an Agent

**Request:**
```
POST /agent/agent123/users/add
Content-Type: application/json

{
  "users": ["alice", "bob"]
}
```

**Response:**
```json
{
  "users": ["alice", "bob"],
  "status": "ok",
  "code": 200
}
```

**Log Entry:**
```
INFO 3304 users added to agent agent123: [alice bob]
```

### Example: Removing Users from an Agent

**Request:**
```
POST /agent/agent123/users/remove
Content-Type: application/json

{
  "users": ["alice"]
}
```

**Response:**
```json
{
  "users": ["bob"],
  "status": "ok",
  "code": 200
}
```

**Log Entry:**
```
INFO 3306 users removed from agent agent123: [alice]
```

---

_Last updated: 2025-04-12_
