//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

// Package global provides global variables and functions for UEMAgent that can be imported by any other package
package global

import (
	"fmt"
	"os"
	"runtime"

	"github.com/UnifyEM/UnifyEM/common/interfaces"
	"github.com/UnifyEM/UnifyEM/common/uconfig"
)

type AgentConfig struct {
	C  interfaces.Config     // Config object
	AC interfaces.Parameters // Agent configuration
	AP interfaces.Parameters // Agent protected config
}

// Config loads the configuration from the registry or file system
func Config() (*AgentConfig, error) {
	var err error
	c := &AgentConfig{}

	// For Windows, use the registry
	if runtime.GOOS == "windows" {
		c.C, err = uconfig.New(uconfig.WithWindowsRegistry(Name))
	} else {
		configFiles := UnixConfigFiles
		c.C, err = uconfig.New(uconfig.WithFindOrCreate(configFiles))
	}

	if err != nil {
		return &AgentConfig{}, err
	}

	// Set constraints, including default values
	c.AC, c.AP = setDefaults(c.C)

	// Update the global lost flag when the config is loaded
	Lost = c.AP.Get(ConfigLost).Bool()

	// Check for a data path
	dPath := c.AP.Get(ConfigAgentDataDir).String()
	if dPath == "" {
		var dSearch []string

		if runtime.GOOS == "windows" {
			dSearch = WindowsDefaultDataPaths
		} else {
			dSearch = UnixDefaultDataPaths
		}

		// Look for a suitable directory
		for _, path := range dSearch {
			// createDir will return true if the directory
			// exists or was successfully created
			if uconfig.CreateDir(path) {
				dPath = path
				break
			}
		}

		// Check for success
		if dPath == "" {
			return &AgentConfig{}, fmt.Errorf("unable to determine or create data directory")
		}

		// Save the path to the config
		c.AP.Set(ConfigAgentDataDir, dPath)
	}

	// Check for logfile and if not set one
	logFile := c.AP.Get(ConfigAgentLogFile).String()
	if logFile == "" {
		lPath := uconfig.CreateSubDir(dPath, "logs")
		if lPath == "" {
			// fall back to the default
			logFile = DefaultLog()
		} else {
			logFile = lPath + string(os.PathSeparator) + LogName + ".log"
		}
		c.AP.Set(ConfigAgentLogFile, logFile)
	}

	// Attempt to checkpoint the config
	_ = c.Checkpoint()
	return c, err
}

func (c *AgentConfig) Checkpoint() error {
	return c.C.Checkpoint()
}
