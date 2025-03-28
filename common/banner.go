//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

package common

import "fmt"

func Banner(program, version string, build int) {
	fmt.Printf("%s version %s (build %d)\n", program, version, build)
	fmt.Printf("Copyright 2024-2025 Tenebris Technologies Inc.\n")
	fmt.Printf("\nLicense:\n")
	fmt.Printf("  This software is licenced under the Apache License, Version 2.0.\n")
	fmt.Printf("  A copy of the license can be found in the LICENSE file.\n")
	fmt.Printf("\nOpen Source:\n")
	fmt.Printf("  This software is open source and relies upon third-party open source\n")
	fmt.Printf("  packages. If you received this software in binary form, please\n")
	fmt.Printf("  refer to the accompanying documentation for full information.\n")
	fmt.Printf("\n")
}
