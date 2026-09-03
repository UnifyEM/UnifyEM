/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package global

import "github.com/UnifyEM/UnifyEM/app"

// Version and Build come from app so existing call sites stay unchanged.
// Version is the parseable release number; Build is the YYYYMMDD day as an int.
var (
	Version   = app.SemVer()
	Build     = app.BuildDate()
	Copyright = app.Copyright()
)

//goland:noinspection GoUnusedConst
const (
	Name            = "UEMCLI"
	Description     = "UnifyEM CLI"
	LongDescription = "UnifyEM command line interface"
)

var ServerURL string
