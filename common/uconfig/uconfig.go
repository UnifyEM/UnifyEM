/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package uconfig

import (
	"fmt"

	"github.com/UnifyEM/UnifyEM/common/interfaces"
	"github.com/UnifyEM/UnifyEM/common/uconfig/params"
)

// Ensure UConfig implements the Config interface
var _ interfaces.Config = (*UConfig)(nil)

// UConfig holds all configuration data
type UConfig struct {
	windowsRegistry    bool                     // Use the Windows registry (ignored on non-Windows systems)
	windowsRegistryKey string                   // Windows registry key (ignored on non-Windows systems)
	file               string                   // Path to configuration file
	Sets               map[string]params.Params `json:"sets"`
}

// Null returns an empty UConfig instance for testing
//
//goland:noinspection GoUnusedExportedFunction
func Null() interfaces.Config {
	return &UConfig{
		windowsRegistry:    false,
		windowsRegistryKey: "",
		file:               "",
		Sets:               make(map[string]params.Params)}
}

// New returns an UConfig instance
//
//goland:noinspection GoUnusedExportedFunction
func New(options ...func(*UConfig) error) (interfaces.Config, error) {
	c := &UConfig{
		windowsRegistry:    false,
		windowsRegistryKey: "",
		file:               "",
		Sets:               make(map[string]params.Params)}

	// Process options (see options.go)
	for _, op := range options {
		err := op(c)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

// Init initializes the configuration data
func (c *UConfig) Init() {
	for key := range c.Sets {
		c.Sets[key] = params.New()
	}
}

// Save the configuration to the specified file or registry
func (c *UConfig) Save(filename string) error {

	if c.windowsRegistry {
		return c.saveRegistry()
	}

	if filename != "" {
		c.file = filename
	}

	if c.file == "" {
		return fmt.Errorf("a filename is required")
	}
	return c.saveFile()
}

// Delete the configuration file
func (c *UConfig) Delete(filename string) error {
	if c.windowsRegistry {
		return nil
	}

	if filename != "" {
		c.file = filename
	}

	if c.file == "" {
		return fmt.Errorf("a filename is required")
	}
	return c.deleteFile()
}

// Load the configuration from the specified file or registry
func (c *UConfig) Load(filename string) error {
	if c.windowsRegistry {
		return c.loadRegistry()
	}

	if filename != "" {
		c.file = filename
	}

	if c.file == "" {
		return fmt.Errorf("a filename is required")
	}
	return c.loadFile()
}

// Checkpoint saves the configuration to the last loaded file
func (c *UConfig) Checkpoint() error {
	if !c.windowsRegistry {
		if c.file == "" {
			return fmt.Errorf("checkpoint requires a loaded configuration")
		}
	}
	return c.Save("")
}

// GetSets returns a pointer to configuration sets as a map
func (c *UConfig) GetSets() map[string]interfaces.Parameters {
	sets := make(map[string]interfaces.Parameters)
	for key, value := range c.Sets {
		sets[key] = &value
	}
	return sets
}

// GetSet returns a pointer to a specific configuration set
func (c *UConfig) GetSet(set string) interfaces.Parameters {
	if value, ok := c.Sets[set]; ok {
		return &value
	}
	return nil
}

func (c *UConfig) NewSet(key string) interfaces.Parameters {
	if _, ok := c.Sets[key]; !ok {
		c.Sets[key] = params.New()
	}
	temp := c.Sets[key]
	return &temp
}
