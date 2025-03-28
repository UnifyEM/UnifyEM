//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

// Package queue provides a simple in-memory queue for messages from agents. It maintains the
// queue within the package and exports functions so that it can be accessed from various parts
// of the application.
package queue

import (
	"github.com/UnifyEM/UnifyEM/common/schema"
)

// Queue holds the channels used as memory queue
var messages chan schema.AgentMessage

// Init the queue with a buffered channel for AgentMessages
func Init(bufferSize int) {
	messages = make(chan schema.AgentMessage, bufferSize)
}

// Add a message to the queue
func Add(msg schema.AgentMessage) {
	messages <- msg
}

// Read is a non-blocking function tht returns an item from the queue
func Read() (schema.AgentMessage, bool) {
	select {
	case msg := <-messages:
		return msg, true
	default:
		return schema.AgentMessage{}, false
	}
}

// Size returns the number of requests currently in the queue
func Size() int {
	return len(messages)
}

// Close closes the queue
//
//goland:noinspection GoUnusedExportedFunction
func Close() {
	close(messages)
}
