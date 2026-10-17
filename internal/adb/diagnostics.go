package adb

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// USBDevice captures properties of a physical USB hardware device reported by OS PnP / enumeration.
type USBDevice struct {
	DeviceID               string `json:"deviceId"`
	Name                   string `json:"name"`
	Manufacturer           string `json:"manufacturer"`
	PNPClass               string `json:"pnpClass"`
	Status                 string `json:"status"`
	ConfigManagerErrorCode int    `json:"configManagerErrorCode"`
}

// USBDetector abstracts OS-level USB hardware device detection.
type USBDetector interface {
	DetectUSBDevices(ctx context.Context) ([]USBDevice, error)
}

// DiagnosticLikelyCause categorizes the most probable state of connection.
type DiagnosticLikelyCause string

const (
	CauseReady            DiagnosticLikelyCause = "ready"
	CauseUnauthorized     DiagnosticLikelyCause = "unauthorized"
	CauseOffline          DiagnosticLikelyCause = "offline"
	CauseUsbDetectedNoAdb DiagnosticLikelyCause = "usb_detected_no_adb"
	CauseNoUsbHardware    DiagnosticLikelyCause = "no_usb_hardware"
	CauseAdbServerFailed  DiagnosticLikelyCause = "adb_server_failed"
)

// ConnectionDiagnosticResult contains comprehensive connection analysis.
type ConnectionDiagnosticResult struct {
	AdbServerRunning   bool                  `json:"adbServerRunning"`
	AdbVersion         string                `json:"adbVersion"`
	UsbHardwarePresent bool                  `json:"usbHardwarePresent"`
	UsbDevices         []USBDevice           `json:"usbDevices"`
	AdbDevices         []DeviceInfo          `json:"adbDevices"`
	LikelyCause        DiagnosticLikelyCause `json:"likelyCause"`
	Headline           string                `json:"headline"`
	PossibleCauses     []string              `json:"possibleCauses"`
	RecommendedSteps   []string              `json:"recommendedSteps"`
}

// DiagnoseConnection performs an evidence-based multi-tier hardware and ADB inspection.
func (m *Manager) DiagnoseConnection(ctx context.Context, detector USBDetector) (*ConnectionDiagnosticResult, error) {
	result := &ConnectionDiagnosticResult{
		AdbServerRunning:   false,
		AdbVersion:         "",
		UsbHardwarePresent: false,
		UsbDevices:         []USBDevice{},
		AdbDevices:         []DeviceInfo{},
		PossibleCauses:     []string{},
		RecommendedSteps:   []string{},
	}

	// 1. Check ADB binary & version
	ctxVer, cancelVer := context.WithTimeout(ctx, 3*time.Second)
	defer cancelVer()

	ver, err := m.CheckVersion(ctxVer)
	if err == nil && ver != "" && ver != "unknown" {
		result.AdbServerRunning = true
		result.AdbVersion = ver
	}

	// 2. Query ADB attached devices via `adb devices -l`
	if result.AdbServerRunning {
		ctxDev, cancelDev := context.WithTimeout(ctx, 4*time.Second)
		defer cancelDev()

		stdout, _, devErr := m.runner.Run(ctxDev, "", "devices", "-l")
		if devErr == nil {
			result.AdbDevices = ParseDevices(string(stdout))
		}
	}

	// 3. Inspect ADB devices state
	for _, dev := range result.AdbDevices {
		if dev.State == StateReady {
			result.LikelyCause = CauseReady
			result.Headline = fmt.Sprintf("Device %s is connected and authorized.", dev.Serial)
			result.RecommendedSteps = []string{"Device is ready for forensic threat analysis."}
			return result, nil
		}
	}

	for _, dev := range result.AdbDevices {
		if dev.State == StateUnauthorized {
			result.LikelyCause = CauseUnauthorized
			result.Headline = fmt.Sprintf("Device %s detected, but USB debugging is not authorized.", dev.Serial)
			result.PossibleCauses = []string{
				"The RSA fingerprint confirmation dialog on the phone screen was not accepted.",
				"The phone screen was locked when the USB cable was connected.",
			}
			result.RecommendedSteps = []string{
				"Unlock your phone screen.",
				"Look for the 'Allow USB debugging?' popup prompt.",
				"Check 'Always allow from this computer' and tap 'Allow'.",
			}
			return result, nil
		}
	}

	for _, dev := range result.AdbDevices {
		if dev.State == StateOffline {
			result.LikelyCause = CauseOffline
			result.Headline = fmt.Sprintf("Device %s detected by ADB, but connection is offline.", dev.Serial)
			result.PossibleCauses = []string{
				"The ADB daemon connection handshake was interrupted.",
				"The USB cable connection is unstable or intermittent.",
			}
			result.RecommendedSteps = []string{
				"Try 'Soft Reconnect' to reset the ADB connection socket.",
				"Unplug and firmly reconnect the USB cable.",
				"Toggle 'USB debugging' OFF and ON in phone Settings > Developer Options.",
			}
			return result, nil
		}
	}

	// 4. No ADB device detected: inspect OS USB hardware layer
	if detector != nil {
		ctxUSB, cancelUSB := context.WithTimeout(ctx, 4*time.Second)
		defer cancelUSB()

		usbList, usbErr := detector.DetectUSBDevices(ctxUSB)
		if usbErr == nil && len(usbList) > 0 {
			result.UsbHardwarePresent = true
			result.UsbDevices = usbList
		}
	}

	if !result.AdbServerRunning {
		result.LikelyCause = CauseAdbServerFailed
		result.Headline = "ADB server is not responding."
		result.PossibleCauses = []string{
			"ADB executable is missing, blocked by antivirus, or corrupted.",
			"Another process or stale daemon is conflicting with the ADB port.",
		}
		result.RecommendedSteps = []string{
			"Run Preflight check to ensure official Google Platform-Tools are installed.",
			"Click 'Restart ADB Server' to restart the daemon cleanly.",
		}
		return result, nil
	}

	if result.UsbHardwarePresent {
		result.LikelyCause = CauseUsbDetectedNoAdb
		result.Headline = "USB device detected by OS, but ADB interface is not active."
		result.PossibleCauses = []string{
			"USB debugging is disabled in Developer Options on your Android phone.",
			"Phone USB configuration is set to 'Charging only' instead of 'File Transfer' (MTP) or 'PTP'.",
			"Windows is missing the OEM USB driver or Android ADB Composite Interface driver.",
		}
		result.RecommendedSteps = []string{
			"Verify 'USB debugging' is turned ON in phone Settings > Developer Options.",
			"Pull down the phone notification shade and change USB mode to 'File Transfer' (MTP).",
			"If Windows Device Manager shows an unknown device or Code 28, install your manufacturer's official USB driver or use Wireless Debugging.",
		}
		return result, nil
	}

	// Default fallback: No hardware USB device detected
	result.LikelyCause = CauseNoUsbHardware
	result.Headline = "No Android USB device detected by the operating system."
	result.PossibleCauses = []string{
		"The USB cable only supports charging (power-only cable without data wires).",
		"The cable is connected to a loose or unpowered USB port/hub.",
		"The Android phone is powered off or locked in deep sleep.",
	}
	result.RecommendedSteps = []string{
		"Use a certified USB data cable (verify that phone storage opens on your PC).",
		"Connect directly to a motherboard USB port on your PC rather than an external hub.",
		"Ensure the phone is powered on and unlocked.",
		"Alternatively, use 'Wireless Debugging' if your phone runs Android 11+ (API 30+).",
	}
	return result, nil
}

// IsAndroidCandidate returns true if the USB device matches common Android vendor IDs, descriptions, or classes.
func IsAndroidCandidate(dev USBDevice) bool {
	idUpper := strings.ToUpper(dev.DeviceID)
	nameLower := strings.ToLower(dev.Name)
	mfgLower := strings.ToLower(dev.Manufacturer)
	classUpper := strings.ToUpper(dev.PNPClass)

	// Exclude generic internal root hubs and host controllers
	if strings.Contains(idUpper, "ROOT_HUB") || strings.Contains(nameLower, "host controller") {
		return false
	}

	// Specific PnP Classes
	if classUpper == "ANDROIDUSBDEVICECLASS" || classUpper == "WPD" {
		return true
	}

	// Common vendor keywords
	androidKeywords := []string{
		"android", "infinix", "transsion", "samsung", "xiaomi", "redmi",
		"oppo", "vivo", "realme", "pixel", "huawei", "honor", "motorola",
		"oneplus", "google", "tecno", "itel", "adb", "mtp",
	}

	for _, kw := range androidKeywords {
		if strings.Contains(nameLower, kw) || strings.Contains(mfgLower, kw) {
			return true
		}
	}

	// Check common Android USB Vendor IDs:
	// Google: 18D1, Samsung: 04E8, Xiaomi: 2717, Transsion/Infinix: 2A70 / 0E8D / 2717,
	// OPPO/Realme: 22D9, Vivo: 2D95, Motorola: 22B8, LG: 1004, HTC: 0BB4, Sony: 0FCE, Asus: 0B05
	vids := []string{
		"VID_18D1", "VID_04E8", "VID_2717", "VID_2A70", "VID_0E8D",
		"VID_22D9", "VID_2D95", "VID_22B8", "VID_1004", "VID_0BB4",
		"VID_0FCE", "VID_0B05",
	}
	for _, vid := range vids {
		if strings.Contains(idUpper, vid) {
			return true
		}
	}

	return false
}
