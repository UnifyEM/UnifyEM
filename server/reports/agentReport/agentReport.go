/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package agentReport

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/server/data"
)

type Report struct{}

func (r *Report) Report(data *data.Data, req schema.ReportRequest) (schema.Report, error) {
	var agents []schema.AgentMeta
	report := schema.NewReport()

	err := data.ForEach(data.BucketAgentMeta, func(key, value []byte) error {
		var agent schema.AgentMeta
		if err := json.Unmarshal(value, &agent); err != nil {
			return fmt.Errorf("error unmarshalling agent data: %w", err)
		}

		agents = append(agents, agent)
		return nil
	})

	if err != nil {
		return report, err
	}

	// Check schema.CmdRequest.Parameters for a format option
	if format, ok := req.Parameters["format"]; ok {
		if format == schema.ReportTypeJSON {
			jsonData, err := json.Marshal(agents)
			if err != nil {
				return report, fmt.Errorf("failed to serialize agent data: %w", err)
			}
			report.Type = schema.ReportTypeJSON
			report.Data = jsonData
			return report, nil
		}
	}

	// Fall back to string format
	var buffer bytes.Buffer
	buffer.WriteString("Agents:\n")
	for _, agent := range agents {
		buffer.WriteString(fmt.Sprintf("%s, %s, %s\n", agent.AgentID, agent.LastSeen, agent.LastIP))
	}
	report.Data = buffer.Bytes()
	report.Type = schema.ReportTypeString
	return report, nil

}
