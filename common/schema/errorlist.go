/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package schema

type ErrorList []ErrorItem

type ErrorItem struct {
	Message string
}

//goland:noinspection GoUnusedExportedFunction
func NewErrorList() ErrorList {
	return ErrorList{}
}

func (e ErrorList) Append(err ErrorItem) ErrorList {
	return append(e, err)
}

func (e ErrorList) AppendMessage(msg string) ErrorList {
	return append(e, ErrorItem{Message: msg})
}
