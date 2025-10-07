//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/UnifyEM/UnifyEM/agent/communications"
	"github.com/UnifyEM/UnifyEM/agent/functions"
	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/agent/install"
	"github.com/UnifyEM/UnifyEM/agent/queues"
	"github.com/UnifyEM/UnifyEM/agent/userdata"
	"github.com/UnifyEM/UnifyEM/agent/userhelper"
	"github.com/UnifyEM/UnifyEM/common"
	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/uemservice"
	"github.com/UnifyEM/UnifyEM/common/ulogger"
)

var conf *global.AgentConfig
var logger interfaces.Logger
var service *uemservice.Service
var communication *communications.Communications
var lastSync int64
var lastStatus int64
var requestQueue *queues.RequestQueue
var responseQueue *queues.ResponseQueue
var userDataListener *userdata.UserDataListener

func main() {

	// Check for version request
	if len(os.Args) == 2 {
		if strings.ToLower(os.Args[1]) == "version" {
			common.Banner(global.Description, global.Version, global.Build)
			exit(0, false)
		}
	}

	// Check for user-helper mode (runs as user, no root required)
	if len(os.Args) > 1 && os.Args[1] == global.UserHelperFlag {
		// Parse collection interval from args if provided
		interval := global.DefaultCollectionInterval
		if len(os.Args) > 3 && os.Args[2] == global.CollectionIntervalFlag {
			if val, err := strconv.Atoi(os.Args[3]); err == nil && val > 0 {
				interval = val
			}
		}
		runUserHelper(interval)
		os.Exit(0)
	}

	// Make sure this program is running with elevated privileges
	err := install.CheckRootPrivileges()
	if err != nil {
		fmt.Printf("Unable to get root or admin privileges: %v\n", err)
		exit(1, true)
	}

	// check for foreground mode
	if len(os.Args) > 1 && strings.ToLower(os.Args[1]) == "foreground" {
		// Start the service in the foreground - used for testing
		startService(true)
	}

	// launch() provides OS-specific functionality
	// It either calls the startService() or console() function below
	launch()
}

// OS-agnostic console mode
func console() int {
	var err error

	fmt.Println("")

	if len(os.Args) < 2 {
		usage()
		return 1
	}

	// Load or create configuration file
	conf, err = global.Config()
	if err != nil {
		fmt.Printf("Fatal config error: %v\n", err)
		return 1
	}

	// Create a logger
	logger, err = newLogger()
	if err != nil {
		fmt.Printf("Fatal logger error: %v\n", err)
		return 1
	}

	switch strings.ToLower(os.Args[1]) {

	case "install":
		if len(os.Args) != 3 {
			fmt.Println("Installation key required")
			usage()
			return 1
		}

		installer := install.New(conf, logger)
		err = installer.Install(os.Args[2])
		if err != nil {
			fmt.Printf("Installation failed: %v\n", err)
			return 1
		}
		fmt.Println("\nService installed successfully")
		_ = conf.Checkpoint()
		return 0

	case "rekey":
		if len(os.Args) != 3 {
			fmt.Println("Installation key required")
			usage()
			return 1
		}

		installer := install.New(conf, logger)
		err = installer.ReKey(os.Args[2])
		if err != nil {
			fmt.Printf("Rekey failed: %v\n", err)
			return 1
		}
		fmt.Println("\nAgent rekeyed successfully")
		_ = conf.Checkpoint()
		return 0

	case "reset":
		// Attempt to stop agent
		installer := install.New(conf, logger)
		err = installer.Stop()
		if err != nil {
			fmt.Printf("\nError stopping agent: %s\n", err.Error())
		}

		// Reset the configuration
		conf.AP.Delete(global.ConfigAgentID)
		conf.AP.Delete(global.ConfigRefreshToken)
		conf.AP.Delete(global.ConfigLost)

		// Checkpoint the configuration
		err = conf.Checkpoint()
		if err != nil {
			fmt.Printf("\nError resetting configuration: %s\n", err.Error())
			return 1
		}
		fmt.Println("\nConfiguration reset successfully")
		_ = conf.Checkpoint()
		return 0

	case "uninstall":
		installer := install.New(conf, logger)
		err = installer.Uninstall()
		if err != nil {
			fmt.Printf("Uninstallation failed: %v\n", err)
			return 1
		}
		fmt.Println("\nService uninstalled successfully")
		return 0

	case "upgrade":
		// Delay for 30 seconds to allow the service to send outstanding messages
		fmt.Println("Waiting 30 seconds for the running service to send outstanding messages...")
		time.Sleep(30 * time.Second)

		// Upgrade the service
		installer := install.New(conf, logger)
		err = installer.Upgrade()
		if err != nil {
			fmt.Printf("Upgrade failed: %v\n", err)
			return 1
		}
		fmt.Println("\nService upgraded successfully")
		return 0

	case "check":
		installer := install.New(conf, logger)
		installer.Check()
		return 0

	default:
		usage()
	}

	// Assume non-normal termination
	return 1
}

func usage() {
	fmt.Printf("Usage: %s <install <key> | rekey <key> | uninstall | upgrade | check | version>\n", os.Args[0])
}

func exit(code int, delay bool) {
	if //goland:noinspection GoBoolExpressions
	delay && global.ConsoleExitDelay > 0 {
		fmt.Printf("\nExiting with code %d in %d seconds...\n\n", code, global.ConsoleExitDelay)
		time.Sleep(global.ConsoleExitDelay * time.Second)
	} else {
		fmt.Printf("\nExiting with code %d\n\n", code)
	}
	os.Exit(code)
}

func startService(optionalArgs ...bool) {
	var err, logErr error
	var foreground bool

	// Check if an optional argument is provided
	if len(optionalArgs) > 0 {
		foreground = optionalArgs[0]
	}

	// Load the configuration and create a logger
	conf, err = global.Config()
	logger, logErr = newLogger()

	// Check for a configuration error
	if err != nil {
		// Check if logger also failed
		if logErr != nil {
			fmt.Printf("Fatal logger error: %s\n", logErr.Error())
		} else {
			logger.Fatalf(8001, "unable to load or create config: %s", err.Error())
		}
		exit(1, false)
	}

	global.Debug = conf.AC.Get(schema.ConfigAgentDebug).Bool()

	if logErr != nil {
		fmt.Printf("error creating logger: %v\n", err)
		// Continue so that a logging issue doesn't prevent updates, etc.
	}

	// Create agent and response queue
	requestQueue = queues.NewRequestQueue(global.TaskQueueSize)
	responseQueue = queues.NewResponseQueue(global.TaskQueueSize)

	// Create a new communication object
	communication, err = communications.New(
		communications.WithLogger(logger),
		communications.WithConfig(conf),
		communications.WithRequestQueue(requestQueue),
		communications.WithResponseQueue(responseQueue))

	if err != nil {
		logger.Fatalf(8002, "unable to create communication object: %s", err.Error())
		exit(1, false)
	}

	// Start user data listener (macOS only)
	if runtime.GOOS == "darwin" {
		userDataListener = userdata.New(logger)
		if err := userDataListener.Start(); err != nil {
			logger.Errorf(8003, "Failed to start user data listener: %v", err)
			// Continue without it - will fall back to existing methods
		} else {
			// Clean stale data periodically
			go func() {
				ticker := time.NewTicker(15 * time.Minute)
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						userDataListener.CleanStaleData(30 * time.Minute)
					}
				}
			}()
		}
	}

	// Check for foreground option
	if foreground {
		simulateService(logger, global.TaskTicker)
	}

	service, err = uemservice.New(
		uemservice.WithServiceName(global.Name),
		uemservice.WithServiceVersion(global.Version),
		uemservice.WithServiceBuild(global.Build),
		uemservice.WithLogger(logger),
		uemservice.WithTaskTicker(global.TaskTicker),
		uemservice.WithBackgroundFunc(nil),
		uemservice.WithTasksFunc(ServiceTasks),
		uemservice.WithStopFunc(ServiceStopping),
		uemservice.WithSEid(8500))

	if err != nil {
		logger.Fatalf(8004, "unable to create service: %v\n", err)
		return
	}

	// Try to tell the server we are starting
	_ = communication.SendMessage(fmt.Sprintf("%s version %s (build %d) starting", global.Name, global.Version, global.Build))

	err = service.Start()
	if err != nil {
		logger.Fatalf(8005, "service failed to start: %s", err.Error())

		// Try to tell the server startup failed
		_ = communication.SendMessage(fmt.Sprintf("%s version %s (build %d) failed to start: %s", global.Name, global.Version, global.Build, err.Error()))
		exit(1, false)
	}
}

// ServiceTasks will be called at the interval specified by TaskTicker
func ServiceTasks(interfaces.Logger) {

	// Get current time in unix format
	now := time.Now().Unix()

	// Check status interval and generate internal agent if required
	if now-lastStatus > conf.AC.Get(schema.ConfigAgentStatusInterval).Int64() {
		// Responses are queued if the server is not available, so don't
		// create a status response if there is already one in the queue
		if !responseQueue.StatusPending() {
			lastStatus = now
			sendStatus()
		}
	}

	// Check in with the server if it has been more than global.SyncInterval seconds or a shorter
	// time period applies
	if syncTime(now - lastSync) {
		lastSync = now
		communication.Sync()
	}

	// Process queued requests
	processRequests()

}

func syncTime(elapsed int64) bool {
	if elapsed > conf.AC.Get(schema.ConfigAgentSyncInterval).Int64() {
		return true
	}

	if responseQueue.Pending() && elapsed > conf.AC.Get(schema.ConfigAgentSyncPending).Int64() {
		return true
	}

	if global.Lost && elapsed > conf.AC.Get(schema.ConfigAgentSyncLost).Int64() {
		return true
	}

	if communication.RetryRequired() && elapsed > conf.AC.Get(schema.ConfigAgentSyncRetry).Int64() {
		return true
	}

	return false
}

// ServiceStopping will be called when the service is stopping
func ServiceStopping(interfaces.Logger) {

	// Stop user data listener
	if userDataListener != nil {
		if err := userDataListener.Stop(); err != nil {
			logger.Errorf(8007, "Error stopping user data listener: %v", err)
		}
	}

	// Save the configuration
	//err := conf.Checkpoint()
	//if err != nil {
	//	logger.Infof(8006, "error saving configuration: %s", err.Error())
	//}

	// Try to tell the server
	_ = communication.SendMessage(fmt.Sprintf("%s version %s (build %d) stopping", global.Name, global.Version, global.Build))
}

// processRequests reads requests from the agent queue and executes them
func processRequests() {

	// Initialize the command functions package
	cmd, err := functions.New(
		functions.WithLogger(logger),
		functions.WithConfig(conf),
		functions.WithComms(communication),
		functions.WithUserDataSource(userDataListener))
	if err != nil {
		logger.Errorf(8050, "error initializing command functions module: %s", err.Error())
		return
	}

	for {
		// Read a request from the request queue
		request, ok := requestQueue.Read()
		if !ok {
			// No more requests in the queue
			return
		}

		// Execute the request
		exeError := executeRequest(cmd, request)
		if exeError != nil {
			// Log the error and continue to the next agent
			logger.Errorf(8052, "error executing request [%s] %s: %s",
				request.RequestID, request.Request, exeError.Error())
		}
	}
}

// executeRequest executes a request using the command functions module
func executeRequest(cmd *functions.Command, request schema.AgentRequest) error {

	logFields := fields.NewFields(
		fields.NewField("request", request.Request),
		fields.NewField("requestID", request.RequestID))

	logger.Info(8051, "executing", logFields)

	response := cmd.ExecuteRequest(request)
	if response.Response != "" {

		// Add the response to the response queue
		responseQueue.Add(response)

		// Log the response
		logFields.Append(
			fields.NewField("success", response.Success),
			fields.NewField("response", response.Response))

		logger.Info(8053, "queued response", logFields)
		return nil
	}
	return errors.New("no response from command")
}

// sendStatus uses the existing status command to generate a response to the server without receiving
// a request. This is useful for sending status information to the server on a regular basis.
func sendStatus() {
	cmd, err := functions.New(
		functions.WithLogger(logger),
		functions.WithConfig(conf),
		functions.WithComms(communication),
		functions.WithUserDataSource(userDataListener))
	if err != nil {
		logger.Errorf(8060, "error initializing command module: %s", err.Error())
		return
	}

	// Get our agentID from the configuration
	agentID := conf.AP.Get(global.ConfigAgentID).String()
	if agentID == "" {
		logger.Warningf(8061, "null AgentID from config, attempting registration")
		communication.Register()
		return
	}

	// Create the request
	request := schema.NewAgentRequest()
	request.AgentID = agentID
	request.RequestID = "status"
	request.Request = "status"
	request.Parameters = make(map[string]string)
	request.Parameters["agent_id"] = agentID

	// Execute the agent
	err = executeRequest(cmd, request)
	if err != nil {
		logger.Errorf(8062, "error executing status request: %s", err.Error())
	}
}

// simulateService runs the service in the foreground for testing. This is particularly useful on Windows.
func simulateService(logger interfaces.Logger, taskTicker time.Duration) {

	logger.Warning(8063, "Running in foreground", nil)

	ticker := time.NewTicker(taskTicker)
	defer ticker.Stop()

	// Channel to listen for termination signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run ServiceTasks at each tick
	go func() {
		for range ticker.C {
			ServiceTasks(logger)
		}
	}()

	// Wait for termination signal
	<-sigChan

	// Call ServiceStopping when the application is terminated
	ServiceStopping(logger)
}

// newLogger creates a new logger on a best-effort basis
func newLogger() (interfaces.Logger, error) {
	var err error
	var l interfaces.Logger

	// Try a full logger first based on the loaded configuration
	debug := conf.AC.Get(schema.ConfigAgentDebug).Bool() || global.Debug
	loggerOptions := []ulogger.Option{
		ulogger.WithPrefix(global.LogName),
		ulogger.WithLogStdout(conf.AC.Get(schema.ConfigAgentLogStdout).Bool()),
		ulogger.WithRetention(conf.AC.Get(schema.ConfigAgentLogRetention).Int()),
		ulogger.WithDebug(debug)}

	var optKey string
	switch runtime.GOOS {
	case "windows":
		optKey = schema.ConfigAgentLogWindowsDisk
		if conf.AC.Get(schema.ConfigAgentLogWindowsEvents).Bool() {
			loggerOptions = append(loggerOptions, ulogger.WithWindowsEvents(true))
		}
	case "darwin":
		optKey = schema.ConfigAgentLogMacOSDisk
	case "linux":
		optKey = schema.ConfigAgentLogLinuxDisk
	default:
		optKey = ""
	}

	if optKey != "" {
		if conf.AC.Get(optKey).Bool() {
			loggerOptions = append(loggerOptions, ulogger.WithLogFile(conf.AP.Get(global.ConfigAgentLogFile).String()))
		}
	}

	// Create the logger
	l, err = ulogger.New(loggerOptions...)
	if err != nil {

		// If that fails, create a console-only logger
		return ulogger.New(
			ulogger.WithPrefix(global.LogName),
			ulogger.WithLogFile(""),
			ulogger.WithLogStdout(true),
			ulogger.WithRetention(0),
			ulogger.WithDebug(true))
	}
	return l, nil
}

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

// getCurrentUsername returns the current user's username
func getCurrentUsername() string {
	u, err := user.Current()
	if err != nil {
		return "unknown"
	}
	return u.Username
}
