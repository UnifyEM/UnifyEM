/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package schema

//goland:noinspection GoUnusedConst
const (
	ReportTypeString = "string"
	ReportTypeJSON   = "json"
	ReportTypeFile   = "file"
)

type Report struct {
	Type string
	Name string
	Data []byte
}

func NewReport() Report {
	return Report{}
}
