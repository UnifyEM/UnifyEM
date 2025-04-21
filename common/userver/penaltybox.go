//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package userver

import (
	"math/rand"
	"time"
)

// PenaltyBox imposes a delay between s.FailedRequestsMin and s.FailedRequestsMax milliseconds
func (s *HServer) PenaltyBox() {
	if s.PenaltyBoxMax == 0 || s.PenaltyBoxMin > s.PenaltyBoxMax {
		return
	}

	var delay int
	if s.PenaltyBoxMin == s.PenaltyBoxMax {
		delay = s.PenaltyBoxMin
	} else {
		delay = s.PenaltyBoxMin + rand.Intn(s.PenaltyBoxMax-s.PenaltyBoxMin)
	}

	time.Sleep(time.Duration(delay) * time.Millisecond)
}
