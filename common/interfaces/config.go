/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package interfaces

// Config defines the methods for configuration management
type Config interface {
	Init()
	Load(string) error
	Save(string) error
	Delete(string) error
	Checkpoint() error
	NewSet(string) Parameters
	GetSets() map[string]Parameters
	GetSet(s string) Parameters
	Dump()
}

type Parameters interface {
	Exists(key string) bool
	Set(key string, value any)
	SetDefault(key string, value any)
	SetConstraint(key string, min, max int, def any)
	SetMap(data map[string]any)
	SetStringMap(data map[string]string)
	Delete(key string)
	Get(key string) ParameterValue
	GetMap() map[string]string
	GetStruct(key string, s any) error
	Serialize() (string, error)
	Deserialize(data string) error
	Dump() (string, error)
}

type ParameterValue interface {
	String() string
	Bytes() []byte
	Int() int
	Int64() int64
	Bool() bool
	Base64() []byte
	SplitMap() map[string]any
	SplitList() []string
}
