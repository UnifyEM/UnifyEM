/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package main

import (
	"fmt"

	"github.com/UnifyEM/UnifyEM/agent/functions/status"
)

func main() {
	fmt.Println("Testing macOS status functions:")

	statusMap := status.CollectStatusData(nil)
	for k, v := range statusMap {
		fmt.Printf("%s: %s\n", k, v)
	}
}
