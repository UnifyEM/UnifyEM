# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

UnifyEM is free, open-source, self-hosted endpoint management software for monitoring, managing, and securing endpoints across Windows, macOS, and Linux. The project is **unreleased** and under active development.

## Components

- **uem-server**: Main server with embedded bbolt database, HTTP API (listens on localhost:8080 by default)
- **uem-agent**: Lightweight cross-platform agent installed on managed endpoints
- **uem-cli**: Command-line interface for administrators
- **uem-webui**: Web interface (future/not yet developed)
- **common**: Shared packages used across all components

## Build Commands

### Building all components (development/testing script)
```bash
./test.sh
```
This builds server, CLI, and agents for all platforms to `bin/` directory.

### Building individual components
```bash
# Server
cd server && go build -o ../bin/uem-server

# CLI
cd cli && go build -o ../bin/uem-cli

# Agent (example for specific platform)
cd agent && GOOS=linux GOARCH=amd64 go build -o ../bin/uem-agent-linux-amd64
```

### Build requirements
- Go 1.24 or later
- govulncheck (install: `go install golang.org/x/vuln/cmd/govulncheck@latest`)

### Production deployment script (Ubuntu server example)
See `uem-build.sh` - stops service, builds, deploys, restarts, and runs govulncheck

## Testing & Quality

```bash
# Run vet
go vet ./...

# Run vulnerability check
govulncheck ./...

# Cross-platform checks (important for agent)
GOOS=windows GOARCH=amd64 go vet ./...
GOOS=darwin GOARCH=arm64 go vet ./...
GOOS=linux GOARCH=amd64 go vet ./...
```

**Always compile code after changes to identify compilation errors, especially for cross-platform conditional code.**

## Architecture

### Communication Model
- All communication is HTTPS initiated by agents/CLI to the server (agent-pull model)
- Server uses NGINX for TLS termination, listening on HTTP localhost
- Agents sync at configurable intervals, pulling pending requests and pushing responses
- Authentication uses JWT tokens (access + refresh tokens)

### Agent-Server Interaction
1. Agent registers with server using installation token (contains FQDN + registration token)
2. Server issues unique agent_id, access_token, and refresh_token
3. Agent stores agent_id and refresh_token in config, keeps access_token in memory
4. On each sync: agent retrieves pending commands and sends queued responses
5. Requests are queued in server DB; responses are queued in agent memory until next sync

### Command Handling Architecture
Commands flow through three layers:
1. **Command Definition** (`common/schema/commands/commands.go`): Defines command name, required/optional args, ack requirements
2. **CLI Command** (`cli/functions/cmd/cmd.go`): Allows commands via `uem-cli cmd <subcommand> key=value...`
3. **Server Validation** (`server/api/cmd.go`): Validates and queues commands for agents
4. **Agent Handler** (`agent/functions/<package>/`): Implements `CmdHandler` interface to execute commands

### Database
- Embedded bbolt (key-value store) in server
- Database location: `/opt/uem-server/` (Linux/macOS) or `C:\ProgramData\uem-server` (Windows)
- Pruning runs every 6 hours via `ServiceTasks()`

### Security Features
- CA certificate pinning to prevent MITM attacks (when enabled)
- SHA256 hash verification for file downloads and upgrades
- Digital signatures for agent requests (in development, can be disabled via `agent/global/global.go`)
- Request signing with server's public key (configurable)

### Agent Protection Mode
`agent/global/global.go` contains `PROTECTED` constant. When `true`, dangerous operations (uninstall, wipe) are disabled. Also contains unsafe development flags (`Unsafe`, `DisableHash`, `DisableSig`) that should be `false` in production.

## Adding New Agent Commands

Follow this exact sequence (see `development.md` for details):

1. **Define command** in `common/schema/commands/commands.go`:
   - Add constant for command name
   - In `init()`, add command with required/optional arguments
   - Use `allArgN(x)` for ordered optional args

2. **Add CLI command** in `cli/functions/cmd/cmd.go`:
   - Add as subcommand to cmd
   - Usage: `uem-cli cmd <subcommand> key=value...`

3. **Recompile uem-server**:
   - Server validates commands in postCmd handler
   - Invalid commands are rejected before queuing

4. **Implement agent handler**:
   - Create package in `agent/functions/<commandname>/`
   - Implement `CmdHandler` interface with `Cmd(schema.AgentRequest) (schema.AgentResponse, error)`
   - Add PROTECTED check for dangerous operations
   - Register in `agent/functions/functions.go` New() function using the constant

5. **Add unique log event IDs** for debugging (every log has unique integer)

## Key Directories

- `agent/functions/`: Individual command handler packages (ping, status, execute, user management, etc.)
- `server/api/`: HTTP API endpoint handlers
- `server/db/`: Database operations
- `server/data/`: Data layer between API and database
- `common/schema/`: Shared data structures (requests, responses, commands, config)
- `common/interfaces/`: Interface definitions (logger, config, cache)

## Configuration

### Server
- Config location: `/etc/uem-server.conf` (Linux/macOS) or registry (Windows)
- Modify via: `uem-cli config server set <key>=<value>`
- Listen URL override: `uem-server listen 127.0.0.1:8080`

### Agent
- Config location: `/etc/uem-agent.conf`, `/usr/local/etc/uem-agent.conf`, or `/var/root/uem-agent.conf` (Unix)
- Windows: Registry + `C:\ProgramData\uem-agent`
- Sync intervals controlled by server

### CLI
- Uses environment variables from `~/.uem` file:
  - `UEM_USER`: Admin username
  - `UEM_PASS`: Admin password
  - `UEM_SERVER`: Server URL (e.g., `https://uem.example.com`)

## Installation Commands

```bash
# Server
./uem-server install           # Install as service
./uem-server admin <user> <pw> # Create super admin (stop service first)
./uem-server uninstall         # Remove service
./uem-server foreground        # Run in foreground (testing)

# Agent
./uem-agent install <server-url>/<reg-token>

# Get registration token
./uem-cli regtoken
./uem-cli regtoken new  # Generate new token
```

## Logging

- Server: `/var/log/uem-server.log` (Unix) or Windows Event Log + `C:\ProgramData\uem-server\uem-server.log`
- Logs rotate daily, retained 30 days (configurable)
- Debug mode set in `server/global/global.go` and `agent/global/global.go`

## Important Notes

- **Requires elevated privileges** (root/admin) for installation and many functions
- Agent currently **not supported on Linux** (Windows/macOS only)
- `main` branch is stable; all other branches are for development
- Hash verification: Run `uem-cli files deploy` after updating agents in HTTP directory to update deployment file hashes
- CA pinning, hash verification, and signature validation can be disabled for development in `agent/global/global.go`
- Every log event uses unique integer IDs for debugging
- Commands can be disabled by commenting out `c.addHandler` line in `agent/functions/functions.go`
