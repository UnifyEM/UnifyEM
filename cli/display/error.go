/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package display

// ErrorWrapper is a simple wrapper for CLI error handling.
// If there is an error, it prints it to the console.
func ErrorWrapper(err error) {
	if err != nil {
		println(err.Error())
	}
}
