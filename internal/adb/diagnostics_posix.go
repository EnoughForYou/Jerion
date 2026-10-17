//go:build !windows

package adb

import (
	"context"
)

// DefaultUSBDetector returns the OS-native USB hardware detector.
func DefaultUSBDetector() USBDetector {
	return &PosixUSBDetector{}
}

// PosixUSBDetector provides fallback detection for non-Windows platforms.
type PosixUSBDetector struct{}

func (d *PosixUSBDetector) DetectUSBDevices(ctx context.Context) ([]USBDevice, error) {
	return []USBDevice{}, nil
}
