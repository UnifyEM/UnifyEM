/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package communications

import (
	"errors"

	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/agent/queues"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

type Communications struct {
	retryRequired bool
	logger        interfaces.Logger
	conf          *global.AgentConfig
	requests      *queues.RequestQueue
	responses     *queues.ResponseQueue
	jwt           string
}

func New(options ...func(*Communications) error) (*Communications, error) {

	// Initialize the triggers
	initTriggers()
	triggerStatus.Lost = global.Lost

	// Create a new Communications instance
	c := &Communications{}
	for _, option := range options {
		err := option(c)
		if err != nil {
			return nil, err
		}
	}

	// Check for mandatory fields
	if c.logger == nil {
		return nil, errors.New("logger is required")
	}

	if c.conf == nil {
		return nil, errors.New("config is required")
	}
	return c, nil
}

func WithLogger(logger interfaces.Logger) func(*Communications) error {
	return func(c *Communications) error {
		if logger == nil {
			return errors.New("logger is nil")
		}
		c.logger = logger
		return nil
	}
}

func WithConfig(config *global.AgentConfig) func(*Communications) error {
	return func(c *Communications) error {
		if config == nil {
			return errors.New("config is nil")
		}
		c.conf = config
		return nil
	}
}

func WithRequestQueue(requests *queues.RequestQueue) func(*Communications) error {
	return func(c *Communications) error {
		if requests == nil {
			return errors.New("requests is nil")
		}
		c.requests = requests
		return nil
	}
}

func WithResponseQueue(responses *queues.ResponseQueue) func(*Communications) error {
	return func(c *Communications) error {
		if responses == nil {
			return errors.New("responses is nil")
		}
		c.responses = responses
		return nil
	}
}

func (c *Communications) RetryRequired() bool {
	if c.retryRequired || c.jwt == "" {
		c.retryRequired = false
		return true
	}
	return false
}
