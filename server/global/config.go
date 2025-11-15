/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package global

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/UnifyEM/UnifyEM/common/interfaces"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/uconfig"
)

type ServerConfig struct {
	C  interfaces.Config     // Config object
	SC interfaces.Parameters // Server configuration
	SP interfaces.Parameters // Server private configuration
	AC interfaces.Parameters // Agent configuration
}

// Config creates the configuration object, sets defaults, and
// loads the configuration from the registry or file system
func Config() (*ServerConfig, error) {
	var err error
	c := &ServerConfig{}

	// For Windows, use the registry
	if runtime.GOOS == "windows" {
		c.C, err = uconfig.New(uconfig.WithWindowsRegistry(Name))
	} else {
		configFiles := UnixConfigFiles
		c.C, err = uconfig.New(uconfig.WithFindOrCreate(configFiles))
	}

	if err != nil {
		return &ServerConfig{}, err
	}

	// Set constraints, including default values
	// SC is the general server configuration set
	// SP is the private server configuration set
	c.SC, c.SP = setDefaults(c.C)
	c.AC = schema.SetAgentDefaults(c.C)

	// Make sure there is a registration token
	regToken := c.SP.Get(ConfigRegToken).String()
	if regToken == "" {
		// Generate one
		regToken, err = GenerateToken()
		if err != nil {
			return &ServerConfig{}, err
		}

		// Save the token
		c.SP.Set(ConfigRegToken, regToken)
	}

	// Check for a data path
	dPath := c.SC.Get(ConfigDataPath).String()
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
			return &ServerConfig{}, fmt.Errorf("unable to determine or create data directory")
		}

		// Save the path to the config
		c.SC.Set(ConfigDataPath, dPath)
	}

	// Make sure there is a database path
	dbPath := c.SC.Get(ConfigDBPath).String()
	if dbPath == "" {
		dbPath = uconfig.CreateSubDir(dPath, "db")
		if dbPath == "" {
			return &ServerConfig{}, fmt.Errorf("unable to create database directory in %s", dPath)
		}

		// Save the path to the config
		c.SC.Set(ConfigDBPath, dbPath)
	}

	// Make sure there is a http (files) path
	fPath := c.SC.Get(ConfigFilesPath).String()
	if fPath == "" {
		fPath = uconfig.CreateSubDir(dPath, "http")
		if fPath == "" {
			return &ServerConfig{}, fmt.Errorf("unable to create http directory in %s", dPath)
		}

		// Save the path to the config
		c.SC.Set(ConfigFilesPath, fPath)
	}

	// Check for logfile and if not set one
	logFile := c.SC.Get(ConfigLogFile).String()
	if logFile == "" {
		lPath := uconfig.CreateSubDir(dPath, "logs")
		if lPath == "" {
			// fall back to the default
			logFile = DefaultLog()
		} else {
			logFile = lPath + string(os.PathSeparator) + LogName + ".log"
		}
		c.SC.Set(ConfigLogFile, logFile)
	}

	// Make sure that critical directories exist
	// They could exist in the config file but have been deleted
	if !uconfig.CreateDir(dPath) {
		return &ServerConfig{}, fmt.Errorf("unable to open or create %s: %w", dPath, err)
	}

	if !uconfig.CreateDir(dbPath) {
		return &ServerConfig{}, fmt.Errorf("unable to open or create %s: %w", dbPath, err)
	}

	if !uconfig.CreateDir(fPath) {
		return &ServerConfig{}, fmt.Errorf("unable to open or create %s: %w", fPath, err)
	}

	// Attempt to checkpoint the config
	err = c.C.Checkpoint()
	if err != nil {
		return &ServerConfig{}, fmt.Errorf("unable to checkpoint config: %w", err)
	}
	return c, err
}

// GenerateToken creates a new random token
func GenerateToken() (string, error) {
	// Create a byte slice to hold the random data
	token := make([]byte, TokenLength)

	// Read random data into the byte slice
	if _, err := io.ReadFull(rand.Reader, token); err != nil {
		return "", err
	}

	// Encode the byte slice in base64
	return base64.URLEncoding.EncodeToString(token), nil
}

func (c *ServerConfig) Checkpoint() error {
	return c.C.Checkpoint()
}
