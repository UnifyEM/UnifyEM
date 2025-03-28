//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

package global

import (
	"github.com/UnifyEM/UnifyEM/common"
)

func Banner() {
	common.Banner(Description, Version, Build)
}
