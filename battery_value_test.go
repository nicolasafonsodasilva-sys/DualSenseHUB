// SPDX-License-Identifier: MIT

package main

import "testing"

func TestDecodeBatteryProperty(t *testing.T) {
	tests := []struct {
		name     string
		propType uint32
		data     []byte
		want     int
		ok       bool
	}{
		{name: "byte", propType: devPropTypeByte, data: []byte{67}, want: 67, ok: true},
		{name: "uint16", propType: devPropTypeUint16, data: []byte{100, 0}, want: 100, ok: true},
		{name: "uint32", propType: devPropTypeUint32, data: []byte{42, 0, 0, 0}, want: 42, ok: true},
		{name: "out of range", propType: devPropTypeByte, data: []byte{255}, ok: false},
		{name: "unknown type", propType: 0x99, data: []byte{50}, ok: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := decodeBatteryProperty(tc.propType, tc.data)
			if ok != tc.ok || got != tc.want {
				t.Fatalf("decodeBatteryProperty(%#x, %v) = (%d, %t); want (%d, %t)", tc.propType, tc.data, got, ok, tc.want, tc.ok)
			}
		})
	}
}
