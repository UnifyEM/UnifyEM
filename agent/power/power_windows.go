/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

// Package power provides functions to interact with the Windows power management API.
package power

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Lazy load PowrProf.dll
var (
	modPowrProf              = windows.NewLazySystemDLL("PowrProf.dll")
	procPowerGetActiveScheme = modPowrProf.NewProc("PowerGetActiveScheme")
	procPowerReadACValue     = modPowrProf.NewProc("PowerReadACValue")
	procPowerReadDCValue     = modPowrProf.NewProc("PowerReadDCValue")
	procLocalFree            = windows.NewLazySystemDLL("Kernel32.dll").NewProc("LocalFree")
)

// GUIDs can be verified using powercfg /Qh
//
//goland:noinspection ALL
var (
	GUID_SUB_NONE = windows.GUID{ // SUB_NONE
		Data1: 0xfea3413e,
		Data2: 0x7e05,
		Data3: 0x4911,
		Data4: [8]byte{0x9a, 0x71, 0x70, 0x03, 0x31, 0xf1, 0xc2, 0x94},
	}

	GUID_CONSOLELOCK = windows.GUID{ // CONSOLELOCK
		Data1: 0x0e796bdb,
		Data2: 0x100d,
		Data3: 0x47d6,
		Data4: [8]byte{0xa2, 0xd5, 0xf7, 0xd2, 0xda, 0xa5, 0x1f, 0x51},
	}

	GUID_VIDEO_SUBGROUP = windows.GUID{ // Display power subgroup
		Data1: 0x7516b95f, Data2: 0xf776, Data3: 0x4464,
		Data4: [8]byte{0x8c, 0x53, 0x06, 0x16, 0x7f, 0x40, 0xcc, 0x99},
	}

	GUID_VIDEO_POWERDOWN_TIMEOUT = windows.GUID{ // Video idle / turn off display setting
		Data1: 0x3c0bc021, Data2: 0xc8a8, Data3: 0x4e07,
		Data4: [8]byte{0xa9, 0x73, 0x6b, 0x14, 0xcb, 0xcb, 0x2b, 0x7e},
	}
)

type ScreenLockInfo struct {
	ConsoleLockAC bool
	ConsoleLockDC bool
	TimeoutAC     uint32
	TimeoutDC     uint32
}

// GetActiveScheme retrieves a pointer to the GUID of the active power scheme.
func GetActiveScheme() (*windows.GUID, error) {
	var guidPtr *windows.GUID
	ret, _, err := procPowerGetActiveScheme.Call(
		0, // userRootPowerKey (NULL)
		uintptr(unsafe.Pointer(&guidPtr)),
	)
	if ret != 0 {
		return nil, fmt.Errorf("PowerGetActiveScheme call failed, error code = %d, err = %v", ret, err)
	}
	return guidPtr, nil
}

// ReadACValue reads the AC (plugged-in) value of a power setting.
func ReadACValue(scheme, subgroup, setting *windows.GUID) (uint32, error) {
	var val uint32
	valSize := uint32(unsafe.Sizeof(val))

	ret, _, err := procPowerReadACValue.Call(
		0,
		uintptr(unsafe.Pointer(scheme)),
		uintptr(unsafe.Pointer(subgroup)),
		uintptr(unsafe.Pointer(setting)),
		0,
		uintptr(unsafe.Pointer(&val)),
		uintptr(unsafe.Pointer(&valSize)),
	)
	if ret != 0 {
		return 0, fmt.Errorf("PowerReadACValue error code = %d, err = %v", ret, err)
	}
	return val, nil
}

// ReadDCValue reads the DC (battery) value of a power setting.
func ReadDCValue(scheme, subgroup, setting *windows.GUID) (uint32, error) {
	var val uint32
	valSize := uint32(unsafe.Sizeof(val))

	ret, _, err := procPowerReadDCValue.Call(
		0,
		uintptr(unsafe.Pointer(scheme)),
		uintptr(unsafe.Pointer(subgroup)),
		uintptr(unsafe.Pointer(setting)),
		0,
		uintptr(unsafe.Pointer(&val)),
		uintptr(unsafe.Pointer(&valSize)),
	)
	if ret != 0 {
		return 0, fmt.Errorf("PowerReadDCValue error code = %d, err = %v", ret, err)
	}
	return val, nil
}

// LocalFree frees memory allocated by some system calls (like PowerGetActiveScheme).
func LocalFree(ptr *windows.GUID) error {
	r0, _, e1 := procLocalFree.Call(uintptr(unsafe.Pointer(ptr)))
	if r0 != 0 {
		return e1
	}
	return nil
}

func GetScreenLockInfo() (ScreenLockInfo, error) {
	ret := ScreenLockInfo{}

	activeScheme, err := GetActiveScheme()
	if err != nil {
		return ret, fmt.Errorf("error getting active scheme: %w", err)

	}
	// Freed after use
	defer func(ptr *windows.GUID) {
		_ = LocalFree(ptr)
	}(activeScheme)

	// Read the AC Console Lock value
	acVal, err := ReadACValue(activeScheme, &GUID_SUB_NONE, &GUID_CONSOLELOCK)
	if err != nil {
		return ret, fmt.Errorf("error reading AC value: %w", err)
	}
	ret.ConsoleLockAC = acVal == 1

	// Read the DC Console Lock value
	dcVal, err := ReadDCValue(activeScheme, &GUID_SUB_NONE, &GUID_CONSOLELOCK)
	if err != nil {
		return ret, fmt.Errorf("error reading DC value: %w", err)
	}
	ret.ConsoleLockDC = dcVal == 1

	// Read the AC display-off timeout
	acTimeout, err := ReadACValue(activeScheme, &GUID_VIDEO_SUBGROUP, &GUID_VIDEO_POWERDOWN_TIMEOUT)
	if err != nil {
		return ret, fmt.Errorf("error reading AC timeout: %w", err)
	}
	ret.TimeoutAC = acTimeout

	// Read the DC display-off timeout
	dcTimeout, err := ReadDCValue(activeScheme, &GUID_VIDEO_SUBGROUP, &GUID_VIDEO_POWERDOWN_TIMEOUT)
	if err != nil {
		return ret, fmt.Errorf("error reading DC timeout: %w", err)
	}
	ret.TimeoutDC = dcTimeout

	return ret, nil
}
