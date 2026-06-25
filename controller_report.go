// SPDX-License-Identifier: MIT

package main

// dualSenseReport is the small, platform-independent result of decoding one
// DualSense HID input report. Keeping the parser independent from Win32 makes
// the protocol handling easy to audit and test on every platform.
type dualSenseReport struct {
	Valid          bool
	Simple         bool
	PSPressed      bool
	BatteryValid   bool
	BatteryPercent int
	Charging       bool
}

// parseDualSenseReport understands the three report layouts used by a
// DualSense/DualSense Edge on Windows:
//   - Bluetooth STANDARD/simple report: report ID 0x01, 10 bytes (sometimes
//     padded by Raw Input to 78 bytes). It contains the PS button but no battery.
//   - USB full report: report ID 0x01, 64 bytes.
//   - Bluetooth enhanced/full report: report ID 0x31, 78 bytes.
func parseDualSenseReport(report []byte) dualSenseReport {
	if len(report) < 2 {
		return dualSenseReport{}
	}

	if report[0] == 0x01 && (len(report) == 10 || len(report) == 78) {
		// In the STANDARD packet, the third buttons byte is report[7]. Bit 0 is
		// the Guide/PS button. This packet deliberately has no battery field.
		if len(report) < 8 {
			return dualSenseReport{}
		}
		return dualSenseReport{
			Valid:     true,
			Simple:    true,
			PSPressed: report[7]&0x01 != 0,
		}
	}

	commonOffset := -1
	switch report[0] {
	case 0x01: // USB full report
		if len(report) >= 54 {
			commonOffset = 1
		}
	case 0x31: // Bluetooth enhanced/full report
		if len(report) >= 55 {
			commonOffset = 2
		}
	}
	if commonOffset < 0 || commonOffset+52 >= len(report) {
		return dualSenseReport{}
	}

	decoded := dualSenseReport{
		Valid:     true,
		PSPressed: report[commonOffset+9]&0x01 != 0,
	}

	status := report[commonOffset+52]
	rawBattery := int(status & 0x0F)
	chargeState := int((status >> 4) & 0x0F)

	switch chargeState {
	case 0x0: // discharging
		decoded.BatteryPercent = clampPercent(rawBattery*10 + 5)
		decoded.Charging = false
		decoded.BatteryValid = true
	case 0x1: // charging
		decoded.BatteryPercent = clampPercent(rawBattery*10 + 5)
		decoded.Charging = true
		decoded.BatteryValid = true
	case 0x2: // fully charged
		decoded.BatteryPercent = 100
		decoded.Charging = true
		decoded.BatteryValid = true
	}

	return decoded
}
