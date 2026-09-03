/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

// Package app holds the application's identity and is the only place any of it
// is defined. Read it through the accessors; do not export the constants.
//
// The version is a source const. Build metadata is injected at link time and
// never replaces it.
package app

import (
	"runtime"
	"strconv"
)

const (
	name      = "UnifyEM"
	tagLine   = "Unified Endpoint Management"
	copyright = "Copyright (c) 2024-2026 Tenebris Technologies Inc."

	// Keep this a single assignment on one line: build tooling reads it.
	version = "0.0.62"
)

// Injected via ldflags. Do not put these in a struct: -X cannot write a field
// and fails silently with exit code 0.
//
//	-X github.com/UnifyEM/UnifyEM/app.gitCommit=<sha8>
//	-X github.com/UnifyEM/UnifyEM/app.buildTime=<rfc3339>
//	-X github.com/UnifyEM/UnifyEM/app.goVersion=<go version>
//	-X github.com/UnifyEM/UnifyEM/app.buildNumber=<utc yyyymmddhhmmss>
var (
	gitCommit   string
	buildTime   string
	goVersion   string
	buildNumber string
)

// Name returns the product name.
func Name() string { return name }

// TagLine returns the one-line product description.
func TagLine() string { return tagLine }

// Copyright returns the copyright notice.
func Copyright() string { return copyright }

// Version returns the display form: "0.0.62+1a2b3c4d [20260902155301]",
// degrading as commit and build number are stamped in. Use SemVer for
// anything another program may parse or compare.
func Version() string {
	v := version
	if gitCommit != "" {
		v += "+" + gitCommit
	}
	if buildNumber != "" {
		v += " [" + buildNumber + "]"
	}
	return v
}

// Build returns the UTC link time as yyyymmddhhmmss, or "" if unstamped.
func Build() string { return buildNumber }

// BuildDate returns the UTC link day as YYYYMMDD (int), or 0 if unstamped.
// For existing int-typed protocols that cannot hold the 14-digit Build stamp.
func BuildDate() int {
	if len(buildNumber) < 8 {
		return 0
	}
	n, err := strconv.Atoi(buildNumber[:8])
	if err != nil {
		return 0
	}
	return n
}

// SemVer returns the bare release number, with no commit or build metadata.
func SemVer() string { return version }

// BuildInfo returns the stamped build timestamp and Go toolchain. The
// toolchain falls back to the running binary; the timestamp stays empty
// if unstamped.
func BuildInfo() (string, string) {
	if goVersion == "" {
		return buildTime, runtime.Version()
	}
	return buildTime, goVersion
}
