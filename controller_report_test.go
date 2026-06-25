// SPDX-License-Identifier: MIT

package main

import "testing"

func TestParseBluetoothStandardTenByteReport(t *testing.T) {
	report := make([]byte, 10)
	report[0] = 0x01
	report[7] = 0x01

	got := parseDualSenseReport(report)
	if !got.Valid || !got.Simple || !got.PSPressed {
		t.Fatalf("unexpected decode: %+v", got)
	}
	if got.BatteryValid {
		t.Fatalf("STANDARD Bluetooth report must not claim an embedded battery field: %+v", got)
	}
}

func TestParseBluetoothStandardPaddedReport(t *testing.T) {
	report := make([]byte, 78)
	report[0] = 0x01
	report[7] = 0x01

	got := parseDualSenseReport(report)
	if !got.Valid || !got.Simple || !got.PSPressed {
		t.Fatalf("unexpected decode: %+v", got)
	}
}

func TestParseUSBFullReport(t *testing.T) {
	report := make([]byte, 64)
	report[0] = 0x01
	report[10] = 0x01 // common offset 1 + buttons offset 9
	report[53] = 0x15 // charging, raw level 5

	got := parseDualSenseReport(report)
	if !got.Valid || got.Simple || !got.PSPressed || !got.BatteryValid || !got.Charging || got.BatteryPercent != 55 {
		t.Fatalf("unexpected decode: %+v", got)
	}
}

func TestParseBluetoothEnhancedReport(t *testing.T) {
	report := make([]byte, 78)
	report[0] = 0x31
	report[11] = 0x01 // common offset 2 + buttons offset 9
	report[54] = 0x24 // full state; low nibble ignored

	got := parseDualSenseReport(report)
	if !got.Valid || got.Simple || !got.PSPressed || !got.BatteryValid || !got.Charging || got.BatteryPercent != 100 {
		t.Fatalf("unexpected decode: %+v", got)
	}
}

func TestRejectUnknownReport(t *testing.T) {
	got := parseDualSenseReport([]byte{0x99, 0x00, 0x00})
	if got.Valid {
		t.Fatalf("unknown report was accepted: %+v", got)
	}
}
