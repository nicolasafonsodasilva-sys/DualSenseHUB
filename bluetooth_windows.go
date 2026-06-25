// SPDX-License-Identifier: MIT
//go:build windows

package main

import (
	"fmt"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	genericRead       = 0x80000000
	genericWrite      = 0x40000000
	fileShareRead     = 0x00000001
	fileShareWrite    = 0x00000002
	openExisting      = 3
	fileAttributeNorm = 0x00000080
	invalidHandle     = ^uintptr(0)
)

var (
	hidDLL                   = syscall.NewLazyDLL("hid.dll")
	procHidDGetFeature       = hidDLL.NewProc("HidD_GetFeature")
	procCreateFileWEnhanced  = kernel32.NewProc("CreateFileW")
	enhancedActivationMu     sync.Mutex
	enhancedActivationByPath = make(map[string]*enhancedActivationState)
)

type enhancedActivationState struct {
	inFlight    bool
	lastAttempt time.Time
	lastSuccess time.Time
}

// requestEnhancedBluetoothMode is called only after a Bluetooth STANDARD
// report was observed. Reading either of these documented DualSense feature
// reports makes the controller start sending full 0x31 Bluetooth reports.
// The device is opened with sharing enabled and the handle is closed
// immediately, so DualSenseHUB does not hold the controller or control output.
func requestEnhancedBluetoothMode(device uintptr, now time.Time) {
	path := getRawInputDeviceName(device)
	if path == "" {
		diagnosticLogf("enhanced activation skipped: empty device path handle=0x%X", device)
		return
	}

	key := strings.ToUpper(path)
	enhancedActivationMu.Lock()
	entry := enhancedActivationByPath[key]
	if entry == nil {
		entry = &enhancedActivationState{}
		enhancedActivationByPath[key] = entry
	}
	// A successful feature read should switch reports almost immediately. If
	// simple reports are still arriving after five seconds, allow another try.
	if entry.inFlight || (!entry.lastAttempt.IsZero() && now.Sub(entry.lastAttempt) < 5*time.Second) {
		enhancedActivationMu.Unlock()
		return
	}
	entry.inFlight = true
	entry.lastAttempt = now
	enhancedActivationMu.Unlock()

	go func(devicePath string, stateEntry *enhancedActivationState) {
		ok, detail := activateDualSenseEnhancedReports(devicePath)
		enhancedActivationMu.Lock()
		stateEntry.inFlight = false
		if ok {
			stateEntry.lastSuccess = time.Now()
		}
		enhancedActivationMu.Unlock()
		diagnosticLogf("enhanced activation result ok=%t detail=%s path=%s", ok, detail, devicePath)
	}(path, entry)
}

func activateDualSenseEnhancedReports(path string) (bool, string) {
	// Some HID stacks permit feature access with read/write, some read-only,
	// and some with zero desired access. All attempts use shared access.
	accessModes := []uint32{
		genericRead | genericWrite,
		genericRead,
		0,
	}

	var failures []string
	for _, access := range accessModes {
		handle, _, openErr := procCreateFileWEnhanced.Call(
			uintptr(unsafe.Pointer(utf16Ptr(path))),
			uintptr(access),
			fileShareRead|fileShareWrite,
			0,
			openExisting,
			fileAttributeNorm,
			0,
		)
		if handle == 0 || handle == invalidHandle {
			failures = append(failures, fmt.Sprintf("CreateFile access=0x%X err=%v", access, openErr))
			continue
		}

		ok, reportDetail := tryDualSenseFeatureReports(handle)
		procCloseHandle.Call(handle)
		if ok {
			return true, fmt.Sprintf("access=0x%X %s", access, reportDetail)
		}
		failures = append(failures, fmt.Sprintf("access=0x%X %s", access, reportDetail))
	}

	return false, strings.Join(failures, "; ")
}

func tryDualSenseFeatureReports(handle uintptr) (bool, string) {
	// Pairing info (0x09, 20 bytes) and firmware info (0x20, 64 bytes) are the
	// feature reports used by SDL to enable enhanced Bluetooth input reports.
	reports := []struct {
		id   byte
		size int
	}{
		{id: 0x09, size: 20},
		{id: 0x20, size: 64},
	}

	var failures []string
	for _, report := range reports {
		buffer := make([]byte, report.size)
		buffer[0] = report.id
		result, _, callErr := procHidDGetFeature.Call(
			handle,
			uintptr(unsafe.Pointer(&buffer[0])),
			uintptr(len(buffer)),
		)
		if result != 0 {
			return true, fmt.Sprintf("HidD_GetFeature id=0x%02X size=%d succeeded", report.id, report.size)
		}
		failures = append(failures, fmt.Sprintf("HidD_GetFeature id=0x%02X size=%d err=%v", report.id, report.size, callErr))
	}
	return false, strings.Join(failures, ", ")
}
