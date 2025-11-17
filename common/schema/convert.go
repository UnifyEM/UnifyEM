/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package schema

import "fmt"

func ConvertMapString(data any) (map[string]string, error) {

	// Assert that data is a map[string]interface{}
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("data is not a map[string]interface{}, got %T", data)
	}

	// Convert the map to map[string]string
	result := make(map[string]string)
	for key, value := range dataMap {
		switch v := value.(type) {
		case string:
			result[key] = v
		case int:
			result[key] = fmt.Sprintf("%d", v)
		case float64:
			result[key] = fmt.Sprintf("%f", v)
		case bool:
			result[key] = fmt.Sprintf("%t", v)
		default:
			result[key] = fmt.Sprintf("%v", v)
		}
	}
	return result, nil
}
