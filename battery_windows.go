// SPDX-License-Identifier: MIT
//go:build windows

package main

import (
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

const (
	digcfPresent     = 0x00000002
	digcfAllClasses  = 0x00000004
	spdrpDeviceDesc  = 0x00000000
	spdrpFriendly    = 0x0000000C
	errorNoMoreItems = 259
)

type winGUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type devPropKey struct {
	FmtID winGUID
	PID   uint32
}

type spDevInfoData struct {
	CbSize    uint32
	ClassGUID winGUID
	DevInst   uint32
	Reserved  uintptr
}

var (
	setupapi = syscall.NewLazyDLL("setupapi.dll")

	procSetupDiGetClassDevsW              = setupapi.NewProc("SetupDiGetClassDevsW")
	procSetupDiEnumDeviceInfo             = setupapi.NewProc("SetupDiEnumDeviceInfo")
	procSetupDiOpenDeviceInfoW            = setupapi.NewProc("SetupDiOpenDeviceInfoW")
	procSetupDiGetDeviceInstanceIdW       = setupapi.NewProc("SetupDiGetDeviceInstanceIdW")
	procSetupDiGetDevicePropertyW         = setupapi.NewProc("SetupDiGetDevicePropertyW")
	procSetupDiGetDeviceRegistryPropertyW = setupapi.NewProc("SetupDiGetDeviceRegistryPropertyW")
	procSetupDiDestroyDeviceInfoList      = setupapi.NewProc("SetupDiDestroyDeviceInfoList")

	batteryLookupMu       sync.Mutex
	cachedBatteryDeviceID string
)

var deviceBatteryLevelKey = devPropKey{
	FmtID: winGUID{
		Data1: 0x104EA319,
		Data2: 0x6EE2,
		Data3: 0x4701,
		Data4: [8]byte{0xBD, 0x47, 0x8D, 0xDB, 0xF4, 0x25, 0xBB, 0xE5},
	},
	PID: 2,
}

func queryDualSenseBatteryPercent() (int, bool) {
	batteryLookupMu.Lock()
	defer batteryLookupMu.Unlock()

	deviceSet, _, _ := procSetupDiGetClassDevsW.Call(0, 0, 0, digcfPresent|digcfAllClasses)
	if deviceSet == 0 || deviceSet == ^uintptr(0) {
		return 0, false
	}
	defer procSetupDiDestroyDeviceInfoList.Call(deviceSet)

	if cachedBatteryDeviceID != "" {
		info := newSPDevInfoData()
		opened, _, _ := procSetupDiOpenDeviceInfoW.Call(
			deviceSet,
			uintptr(unsafe.Pointer(utf16Ptr(cachedBatteryDeviceID))),
			0,
			0,
			uintptr(unsafe.Pointer(&info)),
		)
		if opened != 0 {
			if percent, ok := getBatteryProperty(deviceSet, &info); ok {
				return percent, true
			}
		}
		cachedBatteryDeviceID = ""
	}

	for index := uint32(0); ; index++ {
		info := newSPDevInfoData()
		ok, _, callErr := procSetupDiEnumDeviceInfo.Call(
			deviceSet,
			uintptr(index),
			uintptr(unsafe.Pointer(&info)),
		)
		if ok == 0 {
			if errno, isErrno := callErr.(syscall.Errno); isErrno && uint32(errno) == errorNoMoreItems {
				break
			}
			break
		}

		percent, hasBattery := getBatteryProperty(deviceSet, &info)
		if !hasBattery {
			continue
		}

		instanceID := getDeviceInstanceID(deviceSet, &info)
		if isDualSenseInstanceID(instanceID) || isDualSenseDeviceName(getDeviceName(deviceSet, &info)) {
			cachedBatteryDeviceID = instanceID
			return percent, true
		}
	}

	return 0, false
}

func newSPDevInfoData() spDevInfoData {
	info := spDevInfoData{}
	info.CbSize = uint32(unsafe.Sizeof(info))
	return info
}

func getBatteryProperty(deviceSet uintptr, info *spDevInfoData) (int, bool) {
	var propType uint32
	var required uint32
	var value [16]byte
	ok, _, _ := procSetupDiGetDevicePropertyW.Call(
		deviceSet,
		uintptr(unsafe.Pointer(info)),
		uintptr(unsafe.Pointer(&deviceBatteryLevelKey)),
		uintptr(unsafe.Pointer(&propType)),
		uintptr(unsafe.Pointer(&value[0])),
		uintptr(len(value)),
		uintptr(unsafe.Pointer(&required)),
		0,
	)
	if ok == 0 {
		return 0, false
	}
	length := int(required)
	if length <= 0 || length > len(value) {
		length = len(value)
	}
	return decodeBatteryProperty(propType, value[:length])
}

func getDeviceInstanceID(deviceSet uintptr, info *spDevInfoData) string {
	var buffer [512]uint16
	var required uint32
	ok, _, _ := procSetupDiGetDeviceInstanceIdW.Call(
		deviceSet,
		uintptr(unsafe.Pointer(info)),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(len(buffer)),
		uintptr(unsafe.Pointer(&required)),
	)
	if ok == 0 {
		return ""
	}
	return syscall.UTF16ToString(buffer[:])
}

func getDeviceName(deviceSet uintptr, info *spDevInfoData) string {
	if name := getDeviceRegistryString(deviceSet, info, spdrpFriendly); name != "" {
		return name
	}
	return getDeviceRegistryString(deviceSet, info, spdrpDeviceDesc)
}

func getDeviceRegistryString(deviceSet uintptr, info *spDevInfoData, property uint32) string {
	var regType uint32
	var required uint32
	var buffer [512]uint16
	ok, _, _ := procSetupDiGetDeviceRegistryPropertyW.Call(
		deviceSet,
		uintptr(unsafe.Pointer(info)),
		uintptr(property),
		uintptr(unsafe.Pointer(&regType)),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(len(buffer)*2),
		uintptr(unsafe.Pointer(&required)),
	)
	if ok == 0 {
		return ""
	}
	return syscall.UTF16ToString(buffer[:])
}

func isDualSenseInstanceID(id string) bool {
	upper := strings.ToUpper(id)
	sony := strings.Contains(upper, "VID_054C") ||
		strings.Contains(upper, "VID&054C") ||
		strings.Contains(upper, "VID&0002054C")
	product := strings.Contains(upper, "PID_0CE6") ||
		strings.Contains(upper, "PID&0CE6") ||
		strings.Contains(upper, "PID_0DF2") ||
		strings.Contains(upper, "PID&0DF2")
	return sony && product
}

func isDualSenseDeviceName(name string) bool {
	upper := strings.ToUpper(name)
	return strings.Contains(upper, "DUALSENSE") || strings.Contains(upper, "WIRELESS CONTROLLER")
}
