//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package queues

import (
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/schema/commands"
)

// ResponseQueue holds the channels used as memory queue
type ResponseQueue struct {
	queue         chan schema.AgentResponse // Channel for schema.AgentResponse
	statusPending bool                      // Track if there are status requests waiting to be sent
}

// NewResponseQueue initializes a RequestQueue with a buffered channel for schema.Request
func NewResponseQueue(bufferSize int) *ResponseQueue {
	return &ResponseQueue{
		queue:         make(chan schema.AgentResponse, bufferSize),
		statusPending: false,
	}
}

// Add a response to the queue
func (rq *ResponseQueue) Add(resp schema.AgentResponse) {
	rq.queue <- resp
	if resp.Cmd == commands.Status {
		// Set status pending flag
		rq.statusPending = true
	}
}

// Read is a non-blocking function tht returns an item from the queue
func (rq *ResponseQueue) Read() (schema.AgentResponse, bool) {
	select {
	case resp := <-rq.queue:
		if resp.Cmd == commands.Status {
			// Reset status pending flag
			rq.statusPending = false
		}
		return resp, true
	default:
		// The queue is empty, so there can not be a status pending
		rq.statusPending = false
		return schema.AgentResponse{}, false
	}
}

// ReadAll returns a []schema.AgentResponse of all responses in the queue
func (rq *ResponseQueue) ReadAll() []schema.AgentResponse {
	var responses []schema.AgentResponse
	for {
		resp, ok := rq.Read()
		if !ok {
			break
		}
		responses = append(responses, resp)
	}

	// Reset status pending flag
	rq.statusPending = false

	return responses
}

// ReQueue accepts a []schema.AgentResponse and adds them back to the queue
func (rq *ResponseQueue) ReQueue(responses []schema.AgentResponse) {
	for _, resp := range responses {
		rq.Add(resp)
	}
}

// Size returns the number of requests currently in the queue
func (rq *ResponseQueue) Size() int {
	return len(rq.queue)
}

// Pending returns true if there are responses in the queue
func (rq *ResponseQueue) Pending() bool {
	return len(rq.queue) > 0
}

// Close closes the queue
func (rq *ResponseQueue) Close() {
	close(rq.queue)
}

func (rq *ResponseQueue) StatusPending() bool {
	return rq.statusPending
}
