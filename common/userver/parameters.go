/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package userver

import (
	"net/http"

	"github.com/gorilla/mux"
)

// GetParam retrieves a parameter from the agent URL
//
//goland:noinspection GoUnusedExportedFunction
func GetParam(r *http.Request, param string) string {
	vars := mux.Vars(r)
	if value, ok := vars[param]; ok {
		return value
	}
	return ""
}
