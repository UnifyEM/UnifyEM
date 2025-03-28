//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package status

// Check for running antivirus processes
//
//goland:noinspection SpellCheckingInspection
var antivirusProcesses = []string{
	"Avast",
	"Bitdefender",
	"Norton",
	"McAfee",
	"Sophos",
	"SentinelAgent",
}

// List of common antivirus applications and their installation paths on macOS
//
//goland:noinspection SpellCheckingInspection
var macAntivirusPaths = []string{
	"/Applications/Avast.app",
	"/Applications/Bitdefender.app",
	"/Applications/Norton Security.app",
	"/Applications/McAfee.app",
	"/Applications/Sophos.app",
	"/Applications/SentinelOne.app",
}

//goland:noinspection SpellCheckingInspection
var windowsAntivirusKeys = []string{
	`SOFTWARE\AVAST Software\Avast`,
	`SOFTWARE\AVG\Antivirus`,
	`SOFTWARE\Bitdefender`,
	`SOFTWARE\ESET\ESET Security`,
	`SOFTWARE\KasperskyLab`,
	`SOFTWARE\McAfee`,
	`SOFTWARE\Norton`,
	`SOFTWARE\Sophos`,
	`SOFTWARE\TrendMicro`,
	`SOFTWARE\Sentinel Labs`,
	`SOFTWARE\Microsoft\Windows Defender`,
}
