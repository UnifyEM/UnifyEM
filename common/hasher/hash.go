/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package hasher

import (
	"encoding/base64"

	"github.com/UnifyEM/UnifyEM/common/cache"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

type Hasher struct {
	bytes    []byte // raw bytes returned by the hash function
	cache    interfaces.Cache
	useCache bool
}

type Option func(*Hasher)

// New creates a Data object using the supplied options.
func New(opts ...Option) *Hasher {
	// Initializing cache avoids nil pointer dereference
	r := &Hasher{cache: cache.New(0)}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// WithCache sets the hash cache retention time
func WithCache(s int) Option {
	return func(h *Hasher) {
		h.cache = cache.New(s)
		h.useCache = true
	}
}

func (h *Hasher) Bytes() []byte {
	return h.bytes
}

func (h *Hasher) String() string {
	return string(h.bytes)
}

func (h *Hasher) Base64() string {
	return base64.StdEncoding.EncodeToString(h.bytes)
}

func (h *Hasher) Compare(s string) bool {
	if s == "" {
		return false
	}

	if h.bytes == nil {
		return false
	}

	if len(h.bytes) < 1 {
		return false
	}
	return h.Base64() == s
}
