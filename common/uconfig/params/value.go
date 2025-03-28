//
// Copyright (c) 2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

// Package params implements a simple key/value store with constraints that can be serialized to JSON.
// It
package params

import (
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

// Ensure Value implements the ParameterValue interface
var _ interfaces.ParameterValue = (*Value)(nil)

type Value string

// NewValue is a convenience function that returns a "" as a ParameterValue
func NewValue() interfaces.ParameterValue {
	return Value("")
}

// String converts a Value to a string type
func (v Value) String() string {
	return string(v)
}

// Bytes converts a Value to a byte slice
func (v Value) Bytes() []byte {
	return []byte(v.String())
}

// Int converts a Value to an int type
func (v Value) Int() int {
	i, err := strconv.Atoi(v.String())
	if err != nil {
		return 0
	}
	return i
}

// Int64 converts a Value to an int64 type
func (v Value) Int64() int64 {
	i, err := strconv.ParseInt(v.String(), 10, 64)
	if err != nil {
		return 0
	}
	return i
}

// Bool converts a Value to a bool type
func (v Value) Bool() bool {
	b, err := strconv.ParseBool(v.String())
	if err != nil {
		return false
	}
	return b
}

// Base64 converts a Value to a base64 byte slice
func (v Value) Base64() []byte {
	data, err := base64.StdEncoding.DecodeString(v.String())
	if err != nil {
		return []byte{}
	}
	return data
}

// SplitMap converts a comma-separated Value to a map[string]any for quick lookup
func (v Value) SplitMap() map[string]any {
	m := make(map[string]any)
	parts := strings.Split(v.String(), ",")
	for _, part := range parts {
		m[part] = struct{}{}
	}
	return m
}

// SplitList converts a comma-separated Value to a []string
func (v Value) SplitList() []string {
	parts := strings.Split(v.String(), ",")
	return parts
}
