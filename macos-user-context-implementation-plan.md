# macOS User Context Access Implementation Plan

## Executive Summary

Implement a hybrid LaunchDaemon + LaunchAgent architecture to solve user-context data access issues on macOS. A single binary (`uem-agent`) will operate in two modes based on command-line flags, communicating via Unix domain socket.

**Goal:** Enable accurate collection of user-context data (screen lock settings, user preferences) while maintaining system-level privileges for administrative operations.

**Approach:** Single binary, dual-mode operation with IPC via Unix socket.

---

## Architecture Overview

### Current State
- **LaunchDaemon** at `/Library/LaunchDaemons/com.tenebris.uem-agent.plist`
- Runs as root at system boot
- Cannot reliably access user-context data
- Reports "unknown" for screen lock and related settings

### Target State
- **LaunchDaemon** (root, system boot): Handles server communication, privileged operations, aggregates data
- **LaunchAgent** (user context, per-login): Collects user-specific data, sends to daemon via socket
- **Single binary** at `/usr/local/bin/uem-agent` serves both roles
- **Unix domain socket** at `/var/run/uem-agent.sock` for IPC

### Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│ LaunchAgent (per logged-in user)                            │
│ Command: uem-agent --user-helper --collection-interval 300  │
│                                                              │
│ 1. Check if this user is the console user                   │
│ 2. If console user: Collect user-context data               │
│ 3. Connect to Unix socket: /var/run/uem-agent.sock          │
│ 4. Send JSON payload with user data                         │
│ 5. Sleep for collection interval, repeat                    │
└──────────────────────┬───────────────────────────────────────┘
                       │ Unix Socket IPC
                       ↓
┌─────────────────────────────────────────────────────────────┐
│ LaunchDaemon (root, system-wide)                            │
│ Command: uem-agent (default mode)                           │
│                                                              │
│ 1. Listen on Unix socket for user-helper connections        │
│ 2. Receive user-context data from console user helper       │
│ 3. Collect system-level data (FDE, firewall, etc.)          │
│ 4. Aggregate all data                                       │
│ 5. Sync to UEM server                                       │
└─────────────────────────────────────────────────────────────┘
```

---

## Implementation Phases

### Phase 1: Code Structure & Mode Detection
**Goal:** Add dual-mode capability to single binary

**Files to modify:**
- `agent/nix.go` (or create `agent/main_common.go`)
- `agent/global/global.go`
- `agent/install/install_darwin.go`

**Changes:**

1. **Add mode detection in main()**
```go
// agent/nix.go or agent/darwin.go
func main() {
    // Existing privilege check, version check, etc.

    // NEW: Check for user-helper mode
    if len(os.Args) > 1 && os.Args[1] == "--user-helper" {
        // Parse collection interval from args if provided
        interval := 300 // default 5 minutes
        if len(os.Args) > 3 && os.Args[2] == "--collection-interval" {
            if val, err := strconv.Atoi(os.Args[3]); err == nil && val > 0 {
                interval = val
            }
        }
        runUserHelper(interval)
        os.Exit(0)
    }

    // Existing daemon launch logic
    launch()
}
```

2. **Add global constants**
```go
// agent/global/global.go
const (
    UserHelperFlag          = "--user-helper"
    CollectionIntervalFlag  = "--collection-interval"
    DefaultCollectionInterval = 300 // 5 minutes in seconds
    SocketPath              = "/var/run/uem-agent.sock"
    SocketPerms             = 0666 // Allow user processes to connect
)
```

**Deliverable:** Binary can detect and branch based on `--user-helper` flag

---

### Phase 2: User-Helper Mode Implementation
**Goal:** Create lightweight user-context data collector

**New file:** `agent/userhelper/userhelper.go`

**Structure:**
```go
package userhelper

import (
    "encoding/json"
    "net"
    "time"
    "github.com/UnifyEM/UnifyEM/agent/functions/status"
    "github.com/UnifyEM/UnifyEM/agent/global"
    "github.com/UnifyEM/UnifyEM/common/interfaces"
)

type UserHelper struct {
    logger            interfaces.Logger
    config            *global.AgentConfig
    collectionInterval time.Duration
}

type UserContextData struct {
    Username        string            `json:"username"`
    Timestamp       time.Time         `json:"timestamp"`
    ScreenLock      string            `json:"screen_lock"`
    ScreenLockDelay string            `json:"screen_lock_delay"`
    RawData         map[string]string `json:"raw_data"`
}

func New(logger interfaces.Logger, config *global.AgentConfig, intervalSeconds int) *UserHelper {
    return &UserHelper{
        logger:             logger,
        config:             config,
        collectionInterval: time.Duration(intervalSeconds) * time.Second,
    }
}

// Run is the main loop for user-helper mode
func (h *UserHelper) Run() error {
    username := getCurrentUsername()
    h.logger.Infof(3000, "Starting user-helper mode for user %s with collection interval %v",
        username, h.collectionInterval)

    ticker := time.NewTicker(h.collectionInterval)
    defer ticker.Stop()

    // Send initial data immediately
    if err := h.collectAndSend(); err != nil {
        h.logger.Errorf(3001, "Error collecting initial data: %v", err)
    }

    // Periodic collection
    for range ticker.C {
        if err := h.collectAndSend(); err != nil {
            h.logger.Errorf(3002, "Error collecting periodic data: %v", err)
            // Continue running despite errors
        }
    }

    return nil
}

// collectAndSend gathers user-context data and sends to daemon
func (h *UserHelper) collectAndSend() error {
    // Only collect and send if this is the console user
    if !h.isConsoleUser() {
        h.logger.Debugf(3010, "Not console user, skipping data collection")
        return nil
    }

    data := h.collectUserData()
    return h.sendToDaemon(data)
}

// isConsoleUser checks if the current user is the active console user
func (h *UserHelper) isConsoleUser() bool {
    cmd := exec.Command("/usr/bin/stat", "-f", "%Su", "/dev/console")
    output, err := cmd.Output()
    if err != nil {
        h.logger.Errorf(3011, "Error checking console user: %v", err)
        return false
    }

    consoleUser := strings.TrimSpace(string(output))
    currentUser := getCurrentUsername()

    isConsole := consoleUser == currentUser
    h.logger.Debugf(3012, "Console user: %s, Current user: %s, Is console: %v",
        consoleUser, currentUser, isConsole)

    return isConsole
}

// collectUserData gathers user-specific information
func (h *UserHelper) collectUserData() UserContextData {
    // Create status handler to use existing collection functions
    statusHandler := status.New(h.config, h.logger, nil)

    data := UserContextData{
        Username:  getCurrentUsername(),
        Timestamp: time.Now(),
        RawData:   make(map[string]string),
    }

    // These calls will now work because we're running in user context
    data.ScreenLock, _ = statusHandler.screenLock()
    data.ScreenLockDelay = statusHandler.screenLockDelay()

    // Collect additional user-context data
    data.RawData["last_user"] = statusHandler.lastUser()

    return data
}

// sendToDaemon sends data to daemon via Unix socket
func (h *UserHelper) sendToDaemon(data UserContextData) error {
    conn, err := net.DialTimeout("unix", global.SocketPath, 5*time.Second)
    if err != nil {
        return fmt.Errorf("failed to connect to daemon socket: %w", err)
    }
    defer conn.Close()

    // Set write deadline
    conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

    // Send JSON payload
    encoder := json.NewEncoder(conn)
    if err := encoder.Encode(data); err != nil {
        return fmt.Errorf("failed to send data: %w", err)
    }

    h.logger.Debugf(3003, "Sent user data to daemon: screen_lock=%s, delay=%s",
        data.ScreenLock, data.ScreenLockDelay)

    return nil
}

func getCurrentUsername() string {
    user, err := user.Current()
    if err != nil {
        return "unknown"
    }
    return user.Username
}
```

**Deliverable:** User-helper mode can collect and transmit user-context data

---

### Phase 3: Daemon Socket Listener
**Goal:** Add Unix socket server to daemon mode

**New file:** `agent/userdata/listener.go`

**Structure:**
```go
package userdata

import (
    "encoding/json"
    "net"
    "os"
    "sync"
    "time"
    "github.com/UnifyEM/UnifyEM/agent/global"
    "github.com/UnifyEM/UnifyEM/common/interfaces"
)

type UserDataListener struct {
    logger       interfaces.Logger
    listener     net.Listener
    mu           sync.RWMutex
    consoleUserData UserContextData // Only store console user data
    hasData      bool
    running      bool
}

type UserContextData struct {
    Username        string            `json:"username"`
    Timestamp       time.Time         `json:"timestamp"`
    ScreenLock      string            `json:"screen_lock"`
    ScreenLockDelay string            `json:"screen_lock_delay"`
    RawData         map[string]string `json:"raw_data"`
}

func New(logger interfaces.Logger) *UserDataListener {
    return &UserDataListener{
        logger: logger,
    }
}

// Start begins listening on the Unix socket
func (l *UserDataListener) Start() error {
    // Remove stale socket if it exists
    os.Remove(global.SocketPath)

    listener, err := net.Listen("unix", global.SocketPath)
    if err != nil {
        return fmt.Errorf("failed to create socket listener: %w", err)
    }

    // Set permissions so user processes can connect
    if err := os.Chmod(global.SocketPath, global.SocketPerms); err != nil {
        listener.Close()
        return fmt.Errorf("failed to set socket permissions: %w", err)
    }

    l.listener = listener
    l.running = true
    l.logger.Infof(3100, "User data listener started on %s", global.SocketPath)

    // Start accepting connections in background
    go l.acceptLoop()

    return nil
}

// acceptLoop handles incoming connections
func (l *UserDataListener) acceptLoop() {
    for l.running {
        conn, err := l.listener.Accept()
        if err != nil {
            if l.running {
                l.logger.Errorf(3101, "Error accepting connection: %v", err)
            }
            continue
        }

        // Handle each connection in a goroutine
        go l.handleConnection(conn)
    }
}

// handleConnection processes a single user-helper connection
func (l *UserDataListener) handleConnection(conn net.Conn) {
    defer conn.Close()

    // Set read deadline
    conn.SetReadDeadline(time.Now().Add(10 * time.Second))

    var data UserContextData
    decoder := json.NewDecoder(conn)
    if err := decoder.Decode(&data); err != nil {
        l.logger.Errorf(3102, "Error decoding user data: %v", err)
        return
    }

    // Store the received data (console user only)
    l.mu.Lock()
    l.consoleUserData = data
    l.hasData = true
    l.mu.Unlock()

    l.logger.Debugf(3103, "Received console user data from %s: screen_lock=%s, delay=%s",
        data.Username, data.ScreenLock, data.ScreenLockDelay)
}

// GetConsoleUserData retrieves stored user-context data for the console user
func (l *UserDataListener) GetConsoleUserData() (UserContextData, bool) {
    l.mu.RLock()
    defer l.mu.RUnlock()

    return l.consoleUserData, l.hasData
}

// CleanStaleData removes user data older than the specified duration
func (l *UserDataListener) CleanStaleData(maxAge time.Duration) {
    l.mu.Lock()
    defer l.mu.Unlock()

    if !l.hasData {
        return
    }

    cutoff := time.Now().Add(-maxAge)
    if l.consoleUserData.Timestamp.Before(cutoff) {
        l.logger.Debugf(3104, "Removed stale user data for %s", l.consoleUserData.Username)
        l.hasData = false
        l.consoleUserData = UserContextData{}
    }
}

// Stop closes the listener and cleans up
func (l *UserDataListener) Stop() error {
    l.running = false

    if l.listener != nil {
        if err := l.listener.Close(); err != nil {
            return err
        }
    }

    // Clean up socket file
    os.Remove(global.SocketPath)

    l.logger.Infof(3105, "User data listener stopped")
    return nil
}
```

**Deliverable:** Daemon can receive and store user-context data from helpers

---

### Phase 4: Status Collection Integration
**Goal:** Integrate user-helper data into status reports

**Files to modify:**
- `agent/functions/status/status.go`
- `agent/functions/status/status_darwin.go`

**Changes:**

1. **Add user data listener to status handler**
```go
// agent/functions/status/status.go
type Handler struct {
    config         *global.AgentConfig
    logger         interfaces.Logger
    comms          *communications.Communications
    userDataSource *userdata.UserDataListener // NEW
}

func New(config *global.AgentConfig, logger interfaces.Logger,
         comms *communications.Communications,
         userDataSource *userdata.UserDataListener) *Handler {
    return &Handler{
        config:         config,
        logger:         logger,
        comms:          comms,
        userDataSource: userDataSource,
    }
}
```

2. **Modify macOS screen lock functions to use helper data**
```go
// agent/functions/status/status_darwin.go

func (h *Handler) screenLock() (string, error) {
    // First try to get data from user-helper (console user)
    if h.userDataSource != nil {
        userData, exists := h.userDataSource.GetConsoleUserData()
        if exists && time.Since(userData.Timestamp) < 10*time.Minute {
            h.logger.Debugf(2715, "Using screen lock data from console user helper: %s", userData.ScreenLock)
            return userData.ScreenLock, nil
        }
    }

    // Fallback to existing plist/AppleScript methods
    username := h.lastUser()
    if username == "unknown" {
        return "unknown", fmt.Errorf("could not determine last user")
    }

    enabled, requirePassword, _, err := h.getUserScreenSaverStatus(username)
    if err != nil {
        // Existing fallback logic...
    }

    // Rest of existing logic...
}

func (h *Handler) screenLockDelay() string {
    // First try console user-helper data
    if h.userDataSource != nil {
        userData, exists := h.userDataSource.GetConsoleUserData()
        if exists && time.Since(userData.Timestamp) < 10*time.Minute {
            h.logger.Debugf(2716, "Using screen lock delay from console user helper: %s", userData.ScreenLockDelay)
            return userData.ScreenLockDelay
        }
    }

    // Fallback to existing methods...
}
```

**Deliverable:** Status reports use user-helper data when available, fall back gracefully

---

### Phase 5: LaunchAgent Installation
**Goal:** Install and manage both LaunchDaemon and LaunchAgent plists

**Files to modify:**
- `agent/install/install_darwin.go`

**Changes:**

1. **Add LaunchAgent plist constant**
```go
// agent/install/install_darwin.go
const (
    serviceName       = "uem-agent"
    binaryPath        = "/usr/local/bin"
    daemonPlistPath   = "/Library/LaunchDaemons/com.tenebris.uem-agent.plist"
    agentPlistPath    = "/Library/LaunchAgents/com.tenebris.uem-agent.plist"
)

const agentPlistContent = `
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.tenebris.uem-agent</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/uem-agent</string>
        <string>--user-helper</string>
        <string>--collection-interval</string>
        <string>300</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>EnvironmentVariables</key>
    <dict>
        <key>USER_HELPER_LOG</key>
        <string>/tmp/uem-agent-user-%u.log</string>
    </dict>
    <key>StandardOutPath</key>
    <string>/tmp/uem-agent-user-%u.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/uem-agent-user-%u.log</string>
</dict>
</plist>
`
```

2. **Modify installService() to install both plists**
```go
func (i *Install) installService() error {
    // Existing binary copy and permissions logic...

    // Create daemon plist (existing)
    err = i.createPlist(daemonPlistPath, plistContent)
    if err != nil {
        return err
    }

    // NEW: Create agent plist
    err = i.createPlist(agentPlistPath, agentPlistContent)
    if err != nil {
        return err
    }

    // Load daemon
    cmd := exec.Command("launchctl", "load", daemonPlistPath)
    err = cmd.Run()
    if err != nil {
        return fmt.Errorf("could not load launch daemon: %w", err)
    }

    // Load agent (will start for all current and future user sessions)
    cmd = exec.Command("launchctl", "load", "-w", agentPlistPath)
    err = cmd.Run()
    if err != nil {
        // Non-fatal: agent will load when users log in
        fmt.Printf("Warning: could not load launch agent: %v\n", err)
    }

    return nil
}
```

3. **Modify uninstallService() to remove both**
```go
func (i *Install) uninstallService(removeData bool) error {
    // Unload daemon
    cmd := exec.Command("launchctl", "unload", daemonPlistPath)
    err := cmd.Run()
    // ... existing error handling ...

    // NEW: Unload agent
    cmd = exec.Command("launchctl", "unload", agentPlistPath)
    _ = cmd.Run() // Best effort

    // Remove daemon plist
    err = os.Remove(daemonPlistPath)
    // ... existing error handling ...

    // NEW: Remove agent plist
    err = os.Remove(agentPlistPath)
    if err != nil {
        fmt.Printf("Warning: could not remove agent plist: %v\n", err)
    }

    // Rest of existing logic...
}
```

4. **Update createPlist() to accept path parameter**
```go
func (i *Install) createPlist(path string, content string) error {
    err := os.WriteFile(path, []byte(content), 0644)
    if err != nil {
        return fmt.Errorf("could not write plist file: %w", err)
    }
    fmt.Printf("Plist file created at: %s\n", path)
    return nil
}
```

**Deliverable:** Installation creates both daemon and agent, uninstall removes both

---

### Phase 6: Daemon Startup Integration
**Goal:** Start socket listener when daemon launches

**Files to modify:**
- `agent/nix.go` (or `agent/darwin.go`)

**Changes:**

1. **Initialize user data listener in daemon mode**
```go
// In startService() or ServiceBackground()
func ServiceBackground(logger interfaces.Logger) {
    logger.Infof(2000, "Starting background processes")

    // NEW: Start user data listener (macOS only)
    var userDataListener *userdata.UserDataListener
    if runtime.GOOS == "darwin" {
        userDataListener = userdata.New(logger)
        if err := userDataListener.Start(); err != nil {
            logger.Errorf(2001, "Failed to start user data listener: %v", err)
            // Continue without it - will fall back to existing methods
        }

        // Clean stale data periodically
        go func() {
            ticker := time.NewTicker(15 * time.Minute)
            defer ticker.Stop()
            for range ticker.C {
                userDataListener.CleanStaleData(30 * time.Minute)
            }
        }()
    }

    // Create command handler with user data source
    cmdHandler, err := functions.New(
        functions.WithLogger(logger),
        functions.WithConfig(config),
        functions.WithComms(comms),
        functions.WithUserDataSource(userDataListener), // NEW
    )

    // Rest of existing logic...
}
```

2. **Clean up on shutdown**
```go
func ServiceStopping(logger interfaces.Logger) {
    // NEW: Stop user data listener
    if userDataListener != nil {
        if err := userDataListener.Stop(); err != nil {
            logger.Errorf(2002, "Error stopping user data listener: %v", err)
        }
    }

    // Existing shutdown logic...
}
```

**Deliverable:** Daemon starts socket listener, cleans up on exit

---

### Phase 7: User-Helper Entry Point
**Goal:** Implement runUserHelper() function

**Files to modify:**
- `agent/nix.go` or `agent/darwin.go`

**Changes:**

```go
// runUserHelper is called when --user-helper flag is detected
func runUserHelper(collectionInterval int) {
    // Minimal setup - no privilege escalation needed
    username := getCurrentUsername()

    // Create per-user log file
    logPath := fmt.Sprintf("/tmp/uem-agent-user-%s.log", username)

    // Create logger (log to /tmp since we don't have /var/log access)
    logger, err := ulogger.New(
        ulogger.WithPrefix("uem-agent-user"),
        ulogger.WithLogFile(logPath),
        ulogger.WithLogStdout(false),
        ulogger.WithRetention(7),
        ulogger.WithDebug(global.Debug))

    if err != nil {
        fmt.Printf("Error creating logger: %v\n", err)
        os.Exit(1)
    }

    logger.Infof(3200, "Starting user-helper mode for user: %s (interval: %d seconds)",
        username, collectionInterval)

    // Load minimal config (or use defaults)
    config := &global.AgentConfig{
        // Minimal config needed for status collection
    }

    // Create and run user helper
    helper := userhelper.New(logger, config, collectionInterval)

    if err := helper.Run(); err != nil {
        logger.Errorf(3201, "User-helper error: %v", err)
        os.Exit(1)
    }
}

func getCurrentUsername() string {
    user, err := user.Current()
    if err != nil {
        return "unknown"
    }
    return user.Username
}
```

**Deliverable:** User-helper mode runs independently, collects and transmits data

---

## Testing Strategy

### Unit Tests

1. **Socket communication tests**
   - Test data serialization/deserialization
   - Test connection handling
   - Test error conditions (socket doesn't exist, permission denied, etc.)

2. **User data storage tests**
   - Test console user data storage
   - Test stale data cleanup
   - Test GetConsoleUserData() logic

3. **Mode detection tests**
   - Verify --user-helper flag detection
   - Verify default daemon mode

### Integration Tests

1. **Single user scenario (primary test case)**
   - Install both daemon and agent
   - Verify console user data flows from agent to daemon
   - Verify status reports include user-context data
   - Verify non-console users don't send data

2. **No user logged in scenario**
   - Restart with no active user session
   - Verify daemon continues running
   - Verify status reports gracefully handle missing user data

3. **Upgrade scenario**
   - Install v1, then upgrade to v2
   - Verify both daemon and agent restart correctly
   - Verify no orphaned processes

### Manual Testing Checklist

```bash
# 1. Build binary
cd agent && go build -o ../bin/uem-agent

# 2. Install (creates both daemon and agent)
sudo ./uem-agent install <server-url>/<token>

# 3. Verify both plists exist
ls -la /Library/LaunchDaemons/com.tenebris.uem-agent.plist
ls -la /Library/LaunchAgents/com.tenebris.uem-agent.plist

# 4. Verify socket exists
ls -la /var/run/uem-agent.sock

# 5. Check daemon is running
ps aux | grep uem-agent | grep -v grep
# Should show: root ... /usr/local/bin/uem-agent

# 6. Check user-helper is running
ps aux | grep "uem-agent --user-helper" | grep -v grep
# Should show: <username> ... /usr/local/bin/uem-agent --user-helper --collection-interval 300

# 7. Verify socket communication (send test command)
uem-cli cmd status agent_id=<agent-id>

# 8. Check logs
sudo tail -f /var/log/uem-agent.log                    # Daemon logs
tail -f /tmp/uem-agent-user-$(whoami).log             # User-helper logs

# 9. Test screen lock detection
# Change screen lock settings in System Preferences
# Wait for next collection interval (default 5 minutes)
# Verify status report shows correct values

# 10. Test console user detection
# Log in second user via SSH or Fast User Switching
# Verify only console user helper sends data
# Switch console user, verify new console user sends data

# 11. Test upgrade
sudo uem-agent upgrade
# Verify both processes restart
# Verify no orphaned processes

# 12. Test collection interval
# Modify plist to use different interval (e.g., 60 seconds)
# Reload agent: sudo launchctl unload /Library/LaunchAgents/...plist && sudo launchctl load /Library/LaunchAgents/...plist
# Verify data collected at new interval

# 13. Test uninstall
sudo uem-agent uninstall
# Verify both plists removed
# Verify socket removed
# Verify processes terminated
```

---

## Error Handling & Edge Cases

### Socket Connection Failures

**User-helper cannot connect to daemon:**
- **Cause:** Daemon not running, socket permissions issue
- **Handling:** User-helper logs error, retries on next cycle (no crash)
- **User impact:** Status reports fall back to plist-reading methods

**Daemon socket creation fails:**
- **Cause:** Permission issue, stale socket file
- **Handling:** Daemon removes stale socket and retries, logs error if fails
- **User impact:** User-context data unavailable, fallback methods used

### Multi-User Scenarios

**Console user detection:**
- Multiple user-helpers may be running (one per logged-in user)
- Only console user helper sends data (checked via `/dev/console` ownership)
- Non-console users skip data collection silently
- Daemon only receives and stores console user data

**No user logged in:**
- No user-helper running
- Daemon continues normally
- Status reports show "no_user_session" for user-context fields

### Data Staleness

**User logs out but data remains:**
- Cleanup function runs every 15 minutes
- Data older than 30 minutes is removed
- Status collection checks timestamp, rejects stale data (>10 min old)

### Upgrade Scenarios

**Agent upgrade while user-helper running:**
- Daemon unloads both plists
- Binary replaced
- Both plists reloaded
- Both processes restart with new version

---

## Rollback Plan

### If Implementation Fails

1. **Revert code changes**
   - Git revert or checkout previous commit
   - Rebuild binary with original code

2. **Remove LaunchAgent plist**
   ```bash
   sudo launchctl unload /Library/LaunchAgents/com.tenebris.uem-agent.plist
   sudo rm /Library/LaunchAgents/com.tenebris.uem-agent.plist
   ```

3. **Keep LaunchDaemon unchanged**
   - Original functionality preserved
   - Falls back to plist-reading methods

### Compatibility

**Backward compatibility:**
- Old agents (without user-helper) continue working
- New daemon can run without user-helper (falls back to existing methods)
- No breaking changes to API or data structures

**Forward compatibility:**
- Server doesn't need changes
- CLI doesn't need changes
- Status report schema unchanged (just more accurate data)

---

## Performance Considerations

### Resource Usage

**User-helper:**
- Memory: ~10-15 MB (minimal, same binary loaded)
- CPU: Negligible (sleeps for collection interval between checks)
- Disk: Shared binary, per-user log file in /tmp

**Daemon:**
- Memory: +1-2 MB for socket listener and user data cache
- CPU: Negligible (event-driven socket handling)
- Disk: No change

### Scalability

**Single user:**
- One user-helper, minimal overhead

**Multiple users:**
- One user-helper per logged-in user (but only console user sends data)
- Daemon only receives data from console user
- Non-console helpers use minimal CPU (just check console status periodically)

### Network

**No network overhead:**
- Unix socket is local IPC only
- No additional server communication
- Status reports same size/frequency as before

---

## Security Considerations

### Socket Permissions

**Socket file:** `/var/run/uem-agent.sock`
- Created by daemon (root)
- Permissions: 0666 (world-writable)
- **Risk:** Any local process can send data to socket
- **Mitigation:**
  - Daemon validates data format
  - No privileged operations based on user data
  - User data only used for status reporting (read-only)
  - macOS users already trusted (logged into local machine)

### Process Isolation

**User-helper runs as user:**
- Cannot access daemon's memory or files
- Cannot perform privileged operations
- Can only read user's own preferences
- Communication limited to socket

**Daemon runs as root:**
- Full isolation from user processes
- Socket is only attack surface
- Socket data treated as untrusted input

### Attack Scenarios

**Malicious user sends fake data:**
- **Impact:** Status reports show incorrect screen lock status
- **Severity:** Low (no privilege escalation, no data corruption)
- **Mitigation:** Daemon could validate sender UID via socket credentials (future enhancement)

**Socket DoS attack:**
- **Impact:** Daemon connections backlog, CPU usage
- **Severity:** Low (only affects local machine, no remote impact)
- **Mitigation:** Connection timeout, rate limiting (if needed)

---

## Documentation Updates

### Files to update:

1. **README.md**
   - Update "Known issues" section to remove screen lock limitation
   - Add note about dual-mode agent architecture (optional detail)

2. **CLAUDE.md**
   - Add section on dual-mode architecture
   - Document --user-helper flag
   - Document socket IPC mechanism
   - Update testing procedures

3. **development.md**
   - Add guidance for future macOS-specific features requiring user context
   - Document socket protocol (JSON format)

### New documentation:

**macos-architecture.md** (optional)
- Detailed explanation of LaunchDaemon vs LaunchAgent
- Socket communication protocol
- Multi-user handling
- Troubleshooting guide

---

## Success Criteria

### Functional Requirements
- ✅ Screen lock status reports "yes" or "no" (not "unknown") when user logged in
- ✅ Screen lock delay reports accurate value
- ✅ No regressions in other status fields
- ✅ Single binary distribution and upgrade
- ✅ Graceful handling when no user logged in

### Quality Requirements
- ✅ No memory leaks in socket handling
- ✅ No orphaned processes on upgrade
- ✅ Logs provide clear diagnostic information
- ✅ Error handling prevents crashes

### Operational Requirements
- ✅ Installation process unchanged from user perspective
- ✅ Upgrade process works seamlessly
- ✅ Uninstall cleans up all components
- ✅ Console user detection works correctly
- ✅ Configurable collection interval via command-line argument

---

## Implementation Timeline Estimate

**Phase 1:** Code structure & mode detection - 2-3 hours
**Phase 2:** User-helper implementation - 4-5 hours
**Phase 3:** Daemon socket listener - 4-5 hours
**Phase 4:** Status integration - 3-4 hours
**Phase 5:** LaunchAgent installation - 2-3 hours
**Phase 6:** Daemon startup integration - 2-3 hours
**Phase 7:** User-helper entry point - 1-2 hours

**Testing:** 6-8 hours
**Documentation:** 2-3 hours

**Total:** 26-36 hours of development + testing

---

## Decisions Made

1. **Log file location for user-helper:**
   - **Decision:** `/tmp/uem-agent-user-<username>.log` (per-user log files)
   - Rationale: Avoids permission conflicts, easy to debug per-user

2. **User data retention policy:**
   - **Decision:** 30 minutes (hardcoded initially)
   - Can be made configurable later if needed

3. **Multi-user handling:**
   - **Decision:** Report console user only
   - User-helper checks `/dev/console` ownership before sending data
   - Only active console user sends data to daemon
   - Rationale: Fast User Switching highly unlikely on macOS in enterprise environments

4. **Socket authentication:**
   - **Decision:** No authentication required
   - Rationale: Low risk, minimal impact if exploited (only affects status reporting)

5. **Collection interval:**
   - **Decision:** Configurable via `--collection-interval <seconds>` argument
   - Default: 300 seconds (5 minutes)
   - Set in LaunchAgent plist, can be adjusted per deployment

6. **Installation handling:**
   - **Decision:** Silent handling of agent load failures
   - If no user logged in during install, agent loads on next user login
   - Non-fatal warning displayed during installation

---

## Next Steps

1. **Review this plan** - Get approval before implementation
2. **Create feature branch** - `feature/macos-user-context`
3. **Implement Phase 1** - Mode detection and structure
4. **Test Phase 1** - Verify flag handling works
5. **Proceed through phases** - Implement, test, iterate
6. **Integration testing** - Full end-to-end testing
7. **Documentation** - Update all docs
8. **Code review** - Review before merge to dev
9. **Merge to dev** - Staged rollout
10. **Production testing** - Deploy to test systems before production

---

**Document Version:** 1.0
**Date:** 2025-10-07
**Author:** Implementation plan for UnifyEM macOS user-context access
