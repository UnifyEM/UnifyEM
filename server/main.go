//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	_ "golang.org/x/text" // Make swaggo happy

	"github.com/UnifyEM/UnifyEM/common"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
	"github.com/UnifyEM/UnifyEM/common/null"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/uemservice"
	"github.com/UnifyEM/UnifyEM/common/ulogger"
	"github.com/UnifyEM/UnifyEM/server/api"
	"github.com/UnifyEM/UnifyEM/server/data"
	"github.com/UnifyEM/UnifyEM/server/global"
	"github.com/UnifyEM/UnifyEM/server/install"
	"github.com/UnifyEM/UnifyEM/server/queue"
)

// Swaggo data
// @title UEM-Server
// @version 0.1
// @description Unified Endpoint Management Server
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

var conf *global.ServerConfig
var logger interfaces.Logger
var apiInstance *api.API

func main() {

	// Check for version request
	if len(os.Args) == 2 {
		if strings.ToLower(os.Args[1]) == "version" {
			common.Banner(global.Description, global.Version, global.Build)
			exit(0, false)
		}
	}

	// Make sure this program is running with elevated privileges
	err := install.CheckRootPrivileges()
	if err != nil {
		fmt.Printf("Unable to get root or admin privileges: %v\n", err)
		exit(1, true)
	}

	// Set debug mode to on
	global.Debug = true

	// launch() provides OS-specific functionality and then calls startService() or console() below
	launch()
}

// OS-agnostic console mode
func console() {
	var err error

	fmt.Println("")

	if len(os.Args) < 2 {
		usage()
		return
	}

	// Load or create configuration file
	conf, err = global.Config()
	if err != nil {
		fmt.Printf("Fatal config error: %v\n", err)
		return
	}

	switch strings.ToLower(os.Args[1]) {

	case "admin":
		if len(os.Args) != 4 {
			fmt.Println("Usage: password <username> <password>")
			return
		}

		// Set up data access
		d, dataErr := data.New(conf, null.Logger())
		if dataErr != nil {
			fmt.Printf("Data error: %s\n", dataErr.Error())
			return
		}

		user := os.Args[2]
		pass := os.Args[3]

		// Set the admin user
		err = d.SetAuth(user, pass, schema.RoleSuperAdmin)
		if err != nil {
			fmt.Printf("Error setting admin user: %s\n", err.Error())
			return
		}

		fmt.Printf("Password set for super admin user \"%s\"\n", os.Args[2])
		d.Close()
		return

	case "install":
		installer := install.New(conf)
		err = installer.Install()
		if err != nil {
			fmt.Printf("Installation failed: %v\n", err)
			return
		}
		fmt.Println("\nService installed successfully")

	case "uninstall":
		installer := install.New(conf)
		err = installer.Uninstall()
		if err != nil {
			fmt.Printf("Uninstallation failed: %v\n", err)
			return
		}
		fmt.Println("\nService uninstalled successfully")

	case "upgrade":
		installer := install.New(conf)
		err = installer.Upgrade()
		if err != nil {
			fmt.Printf("Upgrade failed: %v\n", err)
			return
		}
		fmt.Println("\nService upgraded successfully")

	case "check":
		installer := install.New(conf)
		installer.Check()

	case "foreground":
		startService(false)

	case "listen":
		if len(os.Args) != 3 {
			fmt.Println("Usage: listen <address>")
			fmt.Println("Example: uem-server listen 127.0.0.1:8080")
			return
		}

		address := os.Args[2]
		if _, err := net.ResolveTCPAddr("tcp", address); err != nil {
			fmt.Printf("Invalid listen address: %v\n", err)
			return
		}

		global.ListenOverride = address
		startService(false)

	default:
		usage()
	}
}

func usage() {
	fmt.Printf("Usage: %s <install | uninstall | upgrade | check | foreground | listen <address> | admin | version>\n", os.Args[0])
}

func exit(code int, delay bool) {
	if delay {
		fmt.Printf("\nExiting with code %d in %d seconds...\n\n", code, global.ConsoleExitDelay)
		time.Sleep(global.ConsoleExitDelay * time.Second)
	} else {
		fmt.Printf("\nExiting with code %d\n\n", code)
	}
	os.Exit(code)
}

func startService(daemon bool) {
	var err error

	// Load the configuration
	conf, err = global.Config()
	if err != nil {
		// Try to create a logger and write the fatal error
		var loggerErr error
		logger, loggerErr = ulogger.New(
			ulogger.WithPrefix(global.LogName),
			ulogger.WithLogFile(global.DefaultLog()),
			ulogger.WithLogStdout(true),
			ulogger.WithRetention(0),
			ulogger.WithDebug(global.Debug))

		if loggerErr != nil {
			fmt.Printf("Fatal logger error: %s\n", err.Error())
			exit(1, false)
		}
		logger.Fatalf(1001, "unable to load or create config: %s", err.Error())
		exit(1, false)
	}

	// Create a logger using the loaded configuration
	logger, err = ulogger.New(
		ulogger.WithPrefix(global.LogName),
		ulogger.WithLogFile(conf.SC.Get(global.ConfigLogFile).String()),
		ulogger.WithLogStdout(conf.SC.Get(global.ConfigLogStdout).Bool()),
		ulogger.WithRetention(conf.SC.Get(global.ConfigLogRetention).Int()),
		ulogger.WithDebug(global.Debug))

	if err != nil {
		fmt.Printf("error creating logger: %v\n", err)
		// Continue so that a logging issue doesn't prevent updates, etc.
	}

	// Initialize the message queue
	queue.Init(global.MessageQueueSize)

	// Check for foreground option (used for testing)
	if !daemon {
		// This function will never return
		serviceForeground(logger)
	}

	// Start the service
	s, err := uemservice.New(
		uemservice.WithServiceName(global.Name),
		uemservice.WithServiceVersion(global.Version),
		uemservice.WithServiceBuild(global.Build),
		uemservice.WithLogger(logger),
		uemservice.WithTaskTicker(global.TaskTicker),
		uemservice.WithBackgroundFunc(ServiceBackground),
		uemservice.WithTasksFunc(ServiceTasks),
		uemservice.WithStopFunc(ServiceStopping),
		uemservice.WithSEid(1500))

	if err != nil {
		logger.Fatalf(1005, "unable to create service: %s", err.Error())
		exit(1, false)
	}

	//goland:noinspection GoDfaErrorMayBeNotNil
	err = s.Start()
	if err != nil {
		logger.Fatalf(1006, "service failed to start: %s", err.Error())
		exit(1, false)
	}
}

// serviceForeground will run as a foreground service instead of using the service module
// This is intended for testing, primarily on Windows
func serviceForeground(logger interfaces.Logger) {
	logger.Infof(2091, "Starting service in foreground")
	ServiceBackground(logger)

	// Infinite loop with task timer
	for {
		time.Sleep(global.TaskTicker * time.Second)
		logger.Infof(2092, "Running tasks in foreground")
		ServiceTasks(logger)
	}
}

// ServiceBackground will be launched as a goroutine when the service starts
func ServiceBackground(logger interfaces.Logger) {
	logger.Infof(2000, "Starting background processes including API")

	// Start the API
	apiInstance = api.New(conf, logger)
	go apiInstance.Start()
}

// ServiceTasks will be called at the interval specified by TaskTicker
func ServiceTasks(_ interfaces.Logger) {
	// Process any messages in the queue
	if queue.Size() > 0 {
		apiInstance.ProcessMessageQueue()
	}
}

// ServiceStopping is called when the service is about to exit
func ServiceStopping(logger interfaces.Logger) {
	// Close the database
	apiInstance.Close()

	// Save the configuration
	err := conf.C.Checkpoint()
	if err != nil {
		logger.Infof(1007, "error saving configuration: %s", err.Error())
	}
}
