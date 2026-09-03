/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package common

import (
	"fmt"
	"os"

	"github.com/UnifyEM/UnifyEM/app"
)

func Banner(program string) {
	fmt.Fprintf(os.Stderr, "%s %s\n%s\n", program, app.Version(), app.Copyright())
	fmt.Fprintf(os.Stderr, "\nLicense:\n")
	fmt.Fprintf(os.Stderr, "  This software is licenced under the Apache License, Version 2.0.\n")
	fmt.Fprintf(os.Stderr, "  A copy of the license can be found in the LICENSE file.\n")
	fmt.Fprintf(os.Stderr, "\nOpen Source:\n")
	fmt.Fprintf(os.Stderr, "  This software is open source and relies upon third-party open source\n")
	fmt.Fprintf(os.Stderr, "  packages. If you received this software in binary form, please\n")
	fmt.Fprintf(os.Stderr, "  refer to the accompanying documentation for full information.\n")
	fmt.Fprintf(os.Stderr, "\n")
}
