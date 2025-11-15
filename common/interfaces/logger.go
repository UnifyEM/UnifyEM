/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package interfaces

// Logger is an interface that defines required logging functions
// This ensures, among other things, that the os-specific implementations
// are consistent.
type Logger interface {
	Debug(uint32, string, Fields)
	Info(uint32, string, Fields)
	Warning(uint32, string, Fields)
	Error(uint32, string, Fields)
	Fatal(uint32, string, Fields)
	Debugf(uint32, string, ...any)
	Infof(uint32, string, ...any)
	Warningf(uint32, string, ...any)
	Errorf(uint32, string, ...any)
	Fatalf(uint32, string, ...any)
}

// Fields is an interface to decouple the logger from the fields package
type Fields interface {
	ToText() string
	ToPairs() []NVPair
}

// NVPair represents a name-value pair
type NVPair interface {
	Name() string
	Value() any
}
