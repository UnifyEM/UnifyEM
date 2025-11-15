/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package reports

import (
	"errors"

	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/server/data"
	"github.com/UnifyEM/UnifyEM/server/reports/agentReport"
)

type ReportHandler interface {
	Report(data *data.Data, req schema.ReportRequest) (schema.Report, error)
}

var handlers = map[string]ReportHandler{
	"agents": &agentReport.Report{},
}

func Get(data *data.Data, req schema.ReportRequest) (schema.Report, error) {
	handler, exists := handlers[req.Report]
	if !exists {
		return schema.Report{}, errors.New("invalid report: " + req.Report)
	}
	return handler.Report(data, req)
}
