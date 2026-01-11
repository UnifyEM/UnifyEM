/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package params

import (
	"fmt"
	"strconv"
)

// enforceAny accepts an any type and returns a Value. It is primarily used to check Set() data
// for validity. If the value is a string, and it is empty, it returns the default. If it is an
// int or int64, it checks the min and max and it out of range returns the default.
func enforceAny(value any, min int, max int, def Value) Value {
	switch v := value.(type) {

	case string:
		if v == "" {
			return def
		}
		return Value(v)

	case int:
		if min != 0 && v < min {
			return def
		}

		if max != 0 && v > max {
			return def
		}
		return Value(fmt.Sprintf("%d", v))

	case int64:
		if min != 0 && v < int64(min) {
			return def
		}

		if max != 0 && v > int64(max) {
			return def
		}
		return Value(fmt.Sprintf("%d", v))

	default:
		return Value(fmt.Sprintf("%v", v))
	}
}

// enforce accepts an Element and operates on the Value. If the value is an empty string, it checks for
// a default. If the value can be converted to an integer, it checks the min and max
// and if out of range, applies the default.
func enforce(e Element) Value {
	// If empty string, check for default
	if e.Value == "" && e.Default != "" {
		return e.Default
	}

	// Try to convert element.Value to an int
	intValue, err := strconv.Atoi(string(e.Value))
	if err == nil {
		// Value is an integer
		// If either max or max are set, enforce them
		if e.Min != 0 && intValue < e.Min {
			return e.Default
		}

		if e.Max != 0 && intValue > e.Max {
			return e.Default
		}
	}
	return e.Value
}
