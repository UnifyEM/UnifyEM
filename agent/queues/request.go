/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package queues

import (
	"github.com/UnifyEM/UnifyEM/common/schema"
)

// RequestQueue holds the channels used as memory queue
type RequestQueue struct {
	queue chan schema.AgentRequest
}

// NewRequestQueue initializes a RequestQueue with a buffered channel for schema.Request
func NewRequestQueue(bufferSize int) *RequestQueue {
	return &RequestQueue{
		queue: make(chan schema.AgentRequest, bufferSize),
	}
}

// Add a request to the queue
func (rq *RequestQueue) Add(req schema.AgentRequest) {
	rq.queue <- req
}

// Read is a non-blocking function tht returns an item from the queue
func (rq *RequestQueue) Read() (schema.AgentRequest, bool) {
	select {
	case req := <-rq.queue:
		return req, true
	default:
		return schema.AgentRequest{}, false
	}
}

// Size returns the number of requests currently in the queue
func (rq *RequestQueue) Size() int {
	return len(rq.queue)
}

// Close closes the queue
func (rq *RequestQueue) Close() {
	close(rq.queue)
}
