/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package fields

import (
	"fmt"

	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

type Fields struct {
	Fields []Field
}

type Field struct {
	K string
	V any
}

// Name returns the key of the field to implement the NVPair interface
func (f Field) Name() string {
	return f.K
}

// Value returns the value of the field to implement the NVPair interface
func (f Field) Value() any {
	return f.V
}

//goland:noinspection GoUnusedExportedFunction
func NewFields(fields ...Field) *Fields {
	return &Fields{Fields: fields}
}

//goland:noinspection GoUnusedExportedFunction
func (f *Fields) Append(fields ...Field) {
	f.Fields = append(f.Fields, fields...)
}

//goland:noinspection GoUnusedExportedFunction
func (f *Fields) AppendKV(key string, value any) {
	f.Fields = append(f.Fields, Field{K: key, V: value})
}

//goland:noinspection GoUnusedExportedFunction
func (f *Fields) AppendMapAny(m map[string]any) {
	for k, v := range m {
		f.Fields = append(f.Fields, Field{K: k, V: v})
	}
}

//goland:noinspection GoUnusedExportedFunction
func (f *Fields) AppendMapString(m map[string]string) {
	for k, v := range m {
		f.Fields = append(f.Fields, Field{K: k, V: v})
	}
}

//goland:noinspection GoUnusedExportedFunction
func NewField(key string, value any) Field {
	return Field{K: key, V: value}
}

// ToText converts the Fields to a string
func (f *Fields) ToText() string {
	if f == nil {
		return ""
	}

	if len(f.Fields) == 0 {
		return ""
	}

	var text string
	for _, field := range f.Fields {
		text += fmt.Sprintf("%s=%v ", field.K, field.V)
	}

	// Remove trailing space
	if len(text) > 0 {
		text = text[:len(text)-1]
	}

	return text
}

// ToPairs implements the ToPairs method
func (f *Fields) ToPairs() []interfaces.NVPair {
	pairs := make([]interfaces.NVPair, len(f.Fields))
	for i, field := range f.Fields {
		pairs[i] = field
	}
	return pairs
}
