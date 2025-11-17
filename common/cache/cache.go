/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package cache

import (
	"time"

	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

// Instance implements the Cache interface and provides
// a simple in-memory cache for byte slices indexed by string keys.
type Instance struct {
	cache map[string]cacheItem // private cache of file hashes
	ttl   int                  // cache time to live
}

type cacheItem struct {
	bytes   []byte
	created time.Time
}

func New(ttl int) interfaces.Cache {
	return &Instance{
		cache: make(map[string]cacheItem),
		ttl:   ttl}
}

func (c *Instance) Clear() {
	c.cache = make(map[string]cacheItem)
}

func (c *Instance) TTL(ttl int) {
	c.ttl = ttl
}

func (c *Instance) Set(f string, data []byte) {
	c.cache[f] = cacheItem{bytes: data, created: time.Now()}
}

func (c *Instance) Get(f string) []byte {
	v, ok := c.cache[f]
	if ok {
		// Expiration check
		if isOlderThan(v.created, c.ttl) {
			delete(c.cache, f)
			return nil
		}

		// Return valid cache hit
		return v.bytes
	}
	return nil
}

func isOlderThan(t time.Time, s int) bool {
	return time.Since(t) > time.Duration(s)*time.Second
}
