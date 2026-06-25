// SPDX-License-Identifier: MIT
//go:build windows

package main

import "testing"

func TestClampPercent(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{name: "below zero", in: -1, want: 0},
		{name: "zero", in: 0, want: 0},
		{name: "normal", in: 67, want: 67},
		{name: "one hundred", in: 100, want: 100},
		{name: "above one hundred", in: 101, want: 100},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := clampPercent(tc.in); got != tc.want {
				t.Fatalf("clampPercent(%d) = %d; want %d", tc.in, got, tc.want)
			}
		})
	}
}

func TestAllOverlayFramesLoad(t *testing.T) {
	for percent := 0; percent <= 100; percent++ {
		for _, charging := range []bool{false, true} {
			frame, err := loadOverlayFrame(percent, charging)
			if err != nil {
				t.Fatalf("loadOverlayFrame(%d, %t): %v", percent, charging, err)
			}
			if frame.width != overlayWidth || frame.height != overlayHeight {
				t.Fatalf("frame %d charging=%t is %dx%d; want %dx%d", percent, charging, frame.width, frame.height, overlayWidth, overlayHeight)
			}
			if len(frame.pixels) != overlayWidth*overlayHeight*4 {
				t.Fatalf("frame %d charging=%t has %d bytes; want %d", percent, charging, len(frame.pixels), overlayWidth*overlayHeight*4)
			}
		}
	}
}

func TestLowBatteryFrameLoads(t *testing.T) {
	frame, err := loadLowBatteryFrame()
	if err != nil {
		t.Fatalf("loadLowBatteryFrame: %v", err)
	}
	if frame.width != 330 || frame.height != overlayHeight {
		t.Fatalf("low battery frame is %dx%d; want 330x%d", frame.width, frame.height, overlayHeight)
	}
	if len(frame.pixels) != frame.width*frame.height*4 {
		t.Fatalf("low battery frame has %d bytes; want %d", len(frame.pixels), frame.width*frame.height*4)
	}
}
