// SPDX-License-Identifier: MIT

package main

import "encoding/binary"

const (
	devPropTypeByte   = 0x00000003
	devPropTypeUint16 = 0x00000005
	devPropTypeUint32 = 0x00000007
	devPropTypeUint64 = 0x00000009
	devPropTypeMask   = 0x00000FFF
)

func decodeBatteryProperty(propType uint32, data []byte) (int, bool) {
	var value uint64
	switch propType & devPropTypeMask {
	case devPropTypeByte:
		if len(data) < 1 {
			return 0, false
		}
		value = uint64(data[0])
	case devPropTypeUint16:
		if len(data) < 2 {
			return 0, false
		}
		value = uint64(binary.LittleEndian.Uint16(data[:2]))
	case devPropTypeUint32:
		if len(data) < 4 {
			return 0, false
		}
		value = uint64(binary.LittleEndian.Uint32(data[:4]))
	case devPropTypeUint64:
		if len(data) < 8 {
			return 0, false
		}
		value = binary.LittleEndian.Uint64(data[:8])
	default:
		return 0, false
	}
	if value > 100 {
		return 0, false
	}
	return int(value), true
}
