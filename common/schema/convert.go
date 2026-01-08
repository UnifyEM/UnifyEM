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

// ConvertAgentStatusData converts response data to AgentStatusData.
// Supports both the new format (with details/info) and legacy format (flat map).
func ConvertAgentStatusData(data any) (AgentStatusData, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return AgentStatusData{}, fmt.Errorf("data is not a map[string]interface{}, got %T", data)
	}

	result := AgentStatusData{
		Details: make(map[string]string),
		Info:    []string{},
	}

	// Check if this is the new format with "details" key
	if details, hasDetails := dataMap["details"]; hasDetails {
		// New format: extract details map
		if detailsMap, ok := details.(map[string]interface{}); ok {
			for key, value := range detailsMap {
				result.Details[key] = fmt.Sprintf("%v", value)
			}
		}

		// Extract info array if present
		if info, hasInfo := dataMap["info"]; hasInfo {
			if infoArray, ok := info.([]interface{}); ok {
				for _, item := range infoArray {
					if str, ok := item.(string); ok {
						result.Info = append(result.Info, str)
					}
				}
			}
		}
	} else {
		// Legacy format: treat entire map as details
		for key, value := range dataMap {
			switch v := value.(type) {
			case string:
				result.Details[key] = v
			case int:
				result.Details[key] = fmt.Sprintf("%d", v)
			case float64:
				result.Details[key] = fmt.Sprintf("%f", v)
			case bool:
				result.Details[key] = fmt.Sprintf("%t", v)
			default:
				result.Details[key] = fmt.Sprintf("%v", v)
			}
		}
	}

	return result, nil
}
