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
