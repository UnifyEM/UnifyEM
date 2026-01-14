/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

// Package params implements a simple key/value store with constraints that can be serialized to JSON.
// It
package params

import (
	"encoding/json"
	"fmt"

	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

// Ensure Params implements the Parameters interface
var _ interfaces.Parameters = (*Params)(nil)

type Element struct {
	Value   Value `json:"value"`
	Default Value `json:"default"`
	Min     int   `json:"min"`
	Max     int   `json:"max"`
}

type Params struct {
	Data map[string]Element
}

// New returns an initialized Params object
func New() Params {
	return Params{Data: make(map[string]Element)}
}

// Exists checks if a key exists in the Params object
func (p *Params) Exists(key string) bool {
	_, ok := p.Data[key]
	return ok
}

// Set a key/value pair in the Params object
func (p *Params) Set(key string, value any) {
	element, ok := p.Data[key]
	if !ok {
		element = Element{}
	}

	// enforceAny deals with empty strings and out of range ints and returns a string
	element.Value = enforceAny(value, element.Min, element.Max, element.Default)
	p.Data[key] = element
}

// SetDefault sets a default value for a key in the Params object
func (p *Params) SetDefault(key string, value any) {
	element, ok := p.Data[key]
	if !ok {
		element = Element{}
	}
	element.Default = Value(fmt.Sprintf("%v", value))
	p.Data[key] = element
}

// SetConstraint sets a min and max constraint ad a default for a key in the Params object
func (p *Params) SetConstraint(key string, min, max int, def any) {
	element, ok := p.Data[key]
	if !ok {
		element = Element{}
	}
	element.Default = Value(fmt.Sprintf("%v", def))
	element.Min = min
	element.Max = max
	p.Data[key] = element
}

// SetMap sets multiple key/value pairs in the Params object
func (p *Params) SetMap(data map[string]any) {
	for key, value := range data {
		// Use set for constraint enforcement and type conversion
		p.Set(key, value)
	}
}

// SetStringMap sets multiple key/value pairs in the Params object
func (p *Params) SetStringMap(data map[string]string) {
	for key, value := range data {
		// Use set for constraint enforcement and type conversion
		p.Set(key, value)
	}
}

// Delete the value for a key, do not enforce constraints, but leave them in place
func (p *Params) Delete(key string) {
	element, ok := p.Data[key]
	if !ok {
		element = Element{}
	}

	element.Value = ""
	p.Data[key] = element
}

// Get a Value from the Params object
func (p *Params) Get(key string) interfaces.ParameterValue {

	// Get the element if it exists
	element, ok := p.Data[key]
	if !ok {
		return NewValue()
	}

	// Enforce the constraints
	ret := enforce(element)

	// If changed, save it
	if ret != element.Value {
		element.Value = ret
		p.Data[key] = element
	}
	return ret
}

// GetMap converts the Params object to a map[string]string
// Constraints are enforced
func (p *Params) GetMap() map[string]string {
	r := make(map[string]string)
	for key, element := range p.Data {
		r[key] = enforce(element).String()
	}
	return r
}

// GetStruct attempts to retrieve the requested value and deserialize it into the supplied structure
func (p *Params) GetStruct(key string, s any) error {
	value := p.Get(key)

	if value.String() == "" {
		return fmt.Errorf("value for key %s is empty", key)
	}

	err := json.Unmarshal(value.Bytes(), &s)
	if err != nil {
		return fmt.Errorf("failed to deserialize value for key %s: %w", key, err)
	}

	return nil
}

// Serialize the keys and values to a map[string]string
func (p *Params) Serialize() (string, error) {
	tempMap := p.GetMap()
	data, err := json.Marshal(tempMap)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (p *Params) Deserialize(data string) error {
	tempMap := make(map[string]string)
	if err := json.Unmarshal([]byte(data), &tempMap); err != nil {
		return err
	}
	for key, value := range tempMap {
		// Use Set() for constraint enforcement
		p.Set(key, value)
	}
	return nil
}

// Dump the Params object to a JSON string for debugging
// Constraints are not enforce at this point
func (p *Params) Dump() (string, error) {
	data, err := json.MarshalIndent(p.Data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
