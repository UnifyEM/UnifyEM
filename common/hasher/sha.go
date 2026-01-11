/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package hasher

import (
	"crypto/sha256"
	"io"
	"os"
)

func (h *Hasher) SHA256File(f string) *Hasher {
	if f == "" {
		return &Hasher{}
	}

	// If ttl is set, the cache is in use
	if h.useCache {
		b := h.cache.Get(f)
		if b != nil {
			// cache hit
			return &Hasher{bytes: b}
		}
	}

	// cache miss - hash the file
	file, err := os.Open(f)
	if err != nil {
		return &Hasher{}
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	hasher := sha256.New()
	if _, err = io.Copy(hasher, file); err != nil {
		return &Hasher{}
	}
	sum := hasher.Sum(nil)

	// If ttl is set, the cache is in use
	if h.useCache {
		h.cache.Set(f, sum)
	}

	return &Hasher{bytes: sum}
}
