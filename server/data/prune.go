/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package data

import (
	"time"

	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/server/global"
)

// PruneDB removes old data from the database
// It is intended to run as a goroutine and therefore
// logs and handles its own errors. Pruning also deletes any
// bad records that are found, including if deserialization fails
func (d *Data) PruneDB() {
	agentRetention := d.conf.SC.Get(global.ConfigAgentRetention).Int()
	requestRetention := d.conf.SC.Get(global.ConfigRequestRetention).Int()
	eventRetention := d.conf.SC.Get(global.ConfigEventRetention).Int()
	startTime := time.Now()

	d.logger.Info(3000, "Pruning database started", fields.NewFields(
		fields.NewField(global.ConfigAgentRetention, agentRetention),
		fields.NewField(global.ConfigRequestRetention, requestRetention),
		fields.NewField(global.ConfigEventRetention, eventRetention)))

	if agentRetention > 0 {
		d.pruneError(d.database.PruneAgents(agentRetention))
	}

	if requestRetention > 0 {
		d.pruneError(d.database.PruneAgentRequests(requestRetention))
	}

	if eventRetention > 0 {
		d.pruneError(d.database.PruneEvents(eventRetention))
	}

	d.logger.Infof(3001, "Pruning database completed in %.2f seconds", time.Since(startTime).Seconds())
}

func (d *Data) pruneError(err error) {
	if err != nil {
		d.logger.Warningf(3002, "error pruning database: %s", err.Error())
	}
}
