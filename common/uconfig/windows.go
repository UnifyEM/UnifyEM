/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

// Code Windows Only
//go:build windows

package uconfig

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// Windows systems use the registry
const regPrefix = "SOFTWARE\\"

func (c *UConfig) saveRegistry() error {

	// Marshal the configuration to JSON
	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("serialization error: %w", err)
	}

	// Base46 encode the data
	encoded := base64.StdEncoding.EncodeToString(data)

	// Save to the registry
	return c.setRegistry("config", encoded)
}

func (c *UConfig) loadRegistry() error {
	c.Init()

	// Retrieve the data from the registry
	encoded, err := c.getRegistry("config")
	if err != nil {
		// If the registry key isn't found, create it
		if strings.Contains(err.Error(), "key not found") {
			return c.saveRegistry()
		}
		return err
	}

	// Base64 decode the data
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("decoding error: %w", err)
	}

	// Deserialize
	err = json.Unmarshal(data, c)
	if err != nil {
		return fmt.Errorf("deserialization error: %w", err)
	}
	return nil
}

func (c *UConfig) setRegistry(key string, value string) error {

	if c.windowsRegistryKey == "" {
		return fmt.Errorf("windows registry key not set")
	}

	// Create path
	rPath := regPrefix + c.windowsRegistryKey

	// Open the registry key
	rKey, _, err := registry.CreateKey(registry.LOCAL_MACHINE, rPath, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to open registry key %s: %v", rPath, err)
	}

	// Defer closing the key
	defer func(rkey registry.Key) {
		_ = rkey.Close()
	}(rKey)

	err = rKey.SetStringValue(strings.ToLower(key), value)
	if err != nil {
		return fmt.Errorf("failed to set %s registry value: %v", strings.ToLower(key), err)
	}

	return nil
}

func (c *UConfig) getRegistry(key string) (string, error) {

	if c.windowsRegistryKey == "" {
		return "", fmt.Errorf("windows registry key not set")
	}

	// Create path
	rPath := regPrefix + c.windowsRegistryKey

	// Open the registry key
	rKey, _, err := registry.CreateKey(registry.LOCAL_MACHINE, rPath, registry.ALL_ACCESS)
	if err != nil {
		return "", fmt.Errorf("failed to open registry key %s: %v", rPath, err)
	}

	// Defer closing the key
	defer func(rkey registry.Key) {
		_ = rkey.Close()
	}(rKey)

	// Get the key
	value, _, err := rKey.GetStringValue(strings.ToLower(key))
	if err != nil {
		return "", errors.New("key not found")
	}
	return value, nil
}
