//
// Copyright (c) 2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package interfaces

type Cache interface {
	TTL(int)            // Cash time to live in seconds
	Clear()             // Clear the cache
	Set(string, []byte) // Set an item in the cache
	Get(string) []byte  // Get an item from the cache
}
