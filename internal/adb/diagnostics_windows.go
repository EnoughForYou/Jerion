//go:build windows

package adb

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"time"
)

// DefaultUSBDetector returns the OS-native USB hardware detector.
func DefaultUSBDetector() USBDetector {
	return &WindowsUSBDetector{}
}

// WindowsUSBDetector queries connected PnP devices via CIM / PowerShell.
type WindowsUSBDetector struct{}

func (d *WindowsUSBDetector) DetectUSBDevices(ctx context.Context) ([]USBDevice, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	psScript := `Get-CimInstance Win32_PnPEntity | Where-Object { $_.PNPClass -in @('USB','WPD','AndroidUsbDeviceClass') -or $_.Name -match 'android|infinix|samsung|xiaomi|redmi|oppo|vivo|realme|pixel|transsion' } | Select-Object DeviceID, Name, Manufacturer, PNPClass, Status, ConfigManagerErrorCode | ConvertTo-Json -Compress`

	cmd := exec.CommandContext(ctxTimeout, "powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	setHideWindow(cmd)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" || trimmed == "null" {
		return []USBDevice{}, nil
	}

	var devices []USBDevice
	if strings.HasPrefix(trimmed, "[") {
		if err := json.Unmarshal([]byte(trimmed), &devices); err != nil {
			return nil, err
		}
	} else if strings.HasPrefix(trimmed, "{") {
		var single USBDevice
		if err := json.Unmarshal([]byte(trimmed), &single); err != nil {
			return nil, err
		}
		devices = append(devices, single)
	}

	candidates := make([]USBDevice, 0, len(devices))
	for _, dev := range devices {
		if IsAndroidCandidate(dev) {
			candidates = append(candidates, dev)
		}
	}

	return candidates, nil
}
