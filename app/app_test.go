/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package app

import (
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestVersion_IsTheOnlySourceOfTruth(t *testing.T) {
	if version == "" {
		t.Fatal("version must not be empty")
	}
	if version == "dev" {
		t.Fatal("version must be a real release number, not a build-time placeholder")
	}
	if !strings.HasPrefix(Version(), version) {
		t.Fatalf("Version must lead with the release number, got %q", Version())
	}
	if SemVer() != version {
		t.Fatalf("SemVer = %q, want %q", SemVer(), version)
	}
}

func TestVersion_ReleaseAndCommitAreOneToken(t *testing.T) {
	og, ob := gitCommit, buildNumber
	t.Cleanup(func() { gitCommit, buildNumber = og, ob })

	gitCommit, buildNumber = "abc12345", "20260902155301"
	got := Version()

	if strings.Contains(got, "(") {
		t.Fatalf("no parentheses — they invite truncation, got %q", got)
	}
	first := strings.Fields(got)[0]
	want := version + "+abc12345"
	if first != want {
		t.Fatalf("truncating at the first space must still leave the release and the commit: got %q, want %q", first, want)
	}
}

func TestVersion_UsesBuildMetadataSeparator(t *testing.T) {
	old := gitCommit
	t.Cleanup(func() { gitCommit = old })

	gitCommit = "abc12345"
	got := strings.Fields(Version())[0]

	if !strings.Contains(got, "+") {
		t.Fatalf("the commit must attach as SemVer build metadata, got %q", got)
	}
	if strings.Contains(strings.TrimPrefix(got, version), "-") {
		t.Fatalf("a hyphen would make this a pre-release, got %q", got)
	}
	if prefix := strings.SplitN(got, "+", 2)[0]; prefix != version {
		t.Fatalf("everything before the + must be the untouched release number, got %q", prefix)
	}
}

func TestVersion_UnstampedBuildIsBare(t *testing.T) {
	og, ob := gitCommit, buildNumber
	t.Cleanup(func() { gitCommit, buildNumber = og, ob })

	gitCommit, buildNumber = "", ""
	if Version() != version {
		t.Fatalf("Version = %q, want %q", Version(), version)
	}
	if strings.Contains(Version(), "+") {
		t.Fatalf("no dangling separator when nothing is stamped in, got %q", Version())
	}
	if strings.Contains(Version(), "[") {
		t.Fatalf("no empty brackets when no build number is stamped in, got %q", Version())
	}
	if Build() != "" {
		t.Fatalf("Build = %q, want empty", Build())
	}
	if BuildDate() != 0 {
		t.Fatalf("BuildDate = %d, want 0", BuildDate())
	}
}

func TestSemVer_IsPlainSemver(t *testing.T) {
	og, ob := gitCommit, buildNumber
	t.Cleanup(func() { gitCommit, buildNumber = og, ob })

	gitCommit, buildNumber = "abc12345", "20260902155301"

	if SemVer() != version {
		t.Fatalf("SemVer = %q, want %q", SemVer(), version)
	}
	if strings.Contains(SemVer(), "+") || strings.Contains(SemVer(), "[") || strings.Contains(SemVer(), " ") {
		t.Fatalf("the protocol form is a single bare token, got %q", SemVer())
	}
	if Version() == SemVer() {
		t.Fatal("the two forms differ once a build is stamped")
	}
}

func TestIdentityAccessors(t *testing.T) {
	if Name() == "" || TagLine() == "" || Copyright() == "" {
		t.Fatalf("identity accessors must not return empty strings: name=%q tagLine=%q copyright=%q",
			Name(), TagLine(), Copyright())
	}
}

func TestBuildInfo_UsesStampedValues(t *testing.T) {
	ob, og := buildTime, goVersion
	t.Cleanup(func() { buildTime, goVersion = ob, og })

	buildTime, goVersion = "2026-02-20T00:00:00Z", "go1.23.0"

	b, g := BuildInfo()
	if b != "2026-02-20T00:00:00Z" || g != "go1.23.0" {
		t.Fatalf("BuildInfo = (%q, %q), want (%q, %q)", b, g, "2026-02-20T00:00:00Z", "go1.23.0")
	}
}

func TestBuildInfo_FallsBackForToolchainOnly(t *testing.T) {
	ob, og := buildTime, goVersion
	t.Cleanup(func() { buildTime, goVersion = ob, og })

	buildTime, goVersion = "", ""

	b, g := BuildInfo()
	if b != "" {
		t.Fatalf("timestamp must stay empty when unstamped, got %q", b)
	}
	if g != runtime.Version() {
		t.Fatalf("toolchain fallback = %q, want %q", g, runtime.Version())
	}
}

func TestVersion_CarriesTheBuildNumber(t *testing.T) {
	og, ob := gitCommit, buildNumber
	t.Cleanup(func() { gitCommit, buildNumber = og, ob })

	gitCommit, buildNumber = "abc12345", "20260902155301"

	want := version + "+abc12345 [20260902155301]"
	if Version() != want {
		t.Fatalf("Version = %q, want %q", Version(), want)
	}
	if Build() != "20260902155301" {
		t.Fatalf("Build = %q, want %q", Build(), "20260902155301")
	}
	if BuildDate() != 20260902 {
		t.Fatalf("BuildDate = %d, want %d", BuildDate(), 20260902)
	}
}

func TestVersion_BuildNumberWithoutCommit(t *testing.T) {
	og, ob := gitCommit, buildNumber
	t.Cleanup(func() { gitCommit, buildNumber = og, ob })

	gitCommit, buildNumber = "", "20260902155301"

	want := version + " [20260902155301]"
	if Version() != want {
		t.Fatalf("Version = %q, want %q", Version(), want)
	}
	if strings.Contains(Version(), "+") {
		t.Fatalf("no separator for a commit that was never stamped, got %q", Version())
	}
	if BuildDate() != 20260902 {
		t.Fatalf("BuildDate = %d, want %d", BuildDate(), 20260902)
	}
}

func TestBuildDate_RequiresAFullDay(t *testing.T) {
	old := buildNumber
	t.Cleanup(func() { buildNumber = old })

	buildNumber = "202609"
	if BuildDate() != 0 {
		t.Fatalf("truncated stamp must not invent a date, got %d", BuildDate())
	}

	buildNumber = "YYYYMMDDHHMMSS"
	if BuildDate() != 0 {
		t.Fatalf("non-numeric stamp must not invent a date, got %d", BuildDate())
	}
}

func TestBuildNumbersSortLexically(t *testing.T) {
	ordered := []string{
		"20260101000000",
		"20260902094117",
		"20260902155301",
		"20260902155302",
		"20261231235959",
		"20270101000000",
	}
	for i := 1; i < len(ordered); i++ {
		if ordered[i-1] >= ordered[i] {
			t.Fatalf("build numbers must compare in the same order as the clock: %q >= %q", ordered[i-1], ordered[i])
		}
		if len(ordered[i]) != len(ordered[0]) {
			t.Fatalf("every build number must be the same width: %q", ordered[i])
		}
	}
}

func TestBuildNumberShapeMatchesTheMakefile(t *testing.T) {
	stamped := time.Date(2026, 9, 2, 15, 53, 1, 0, time.UTC).Format("20060102150405")
	if stamped != "20260902155301" {
		t.Fatalf("Makefile stamp format produced %q, want %q", stamped, "20260902155301")
	}
	if len(stamped) != 14 {
		t.Fatalf("seconds resolution expected length 14, got %d", len(stamped))
	}
}
