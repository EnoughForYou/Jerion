package adb_test

import (
	"context"
	"testing"

	"github.com/EnoughForYou/Jerion/internal/adb"
)

func TestIsAndroidCandidate(t *testing.T) {
	tests := []struct {
		name     string
		device   adb.USBDevice
		expected bool
	}{
		{
			name: "Google Pixel by Vendor ID",
			device: adb.USBDevice{
				DeviceID:     `USB\VID_18D1&PID_4EE7\123456`,
				Name:         "Pixel 7",
				Manufacturer: "Google",
				PNPClass:     "USB",
			},
			expected: true,
		},
		{
			name: "Infinix Note 11S by keyword",
			device: adb.USBDevice{
				DeviceID:     `USB\VID_2717&PID_FF40\0789456`,
				Name:         "Infinix Note 11S",
				Manufacturer: "Infinix",
				PNPClass:     "WPD",
			},
			expected: true,
		},
		{
			name: "Samsung MTP device",
			device: adb.USBDevice{
				DeviceID:     `USB\VID_04E8&PID_6860\SAMSUNG123`,
				Name:         "SAMSUNG Mobile USB Composite Device",
				Manufacturer: "SAMSUNG",
				PNPClass:     "USB",
			},
			expected: true,
		},
		{
			name: "AndroidUsbDeviceClass PnP class",
			device: adb.USBDevice{
				DeviceID:     `USB\OTHER_ID\XYZ`,
				Name:         "ADB Device Interface",
				Manufacturer: "Unknown",
				PNPClass:     "AndroidUsbDeviceClass",
			},
			expected: true,
		},
		{
			name: "Internal Root Hub (negative)",
			device: adb.USBDevice{
				DeviceID:     `USB\ROOT_HUB30\5&123`,
				Name:         "USB Root Hub (USB 3.0)",
				Manufacturer: "(Standard USB HUBs)",
				PNPClass:     "USB",
			},
			expected: false,
		},
		{
			name: "Generic Mouse (negative)",
			device: adb.USBDevice{
				DeviceID:     `USB\VID_1532&PID_006E\RAZER_MOUSE`,
				Name:         "Razer DeathAdder",
				Manufacturer: "Razer",
				PNPClass:     "HIDClass",
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := adb.IsAndroidCandidate(tc.device)
			if actual != tc.expected {
				t.Errorf("IsAndroidCandidate() = %v, want %v", actual, tc.expected)
			}
		})
	}
}

func TestDiagnoseConnection_TableDriven(t *testing.T) {
	tests := []struct {
		name              string
		versionOutput     string
		versionErr        error
		devicesOutput     string
		usbDevices        []adb.USBDevice
		wantLikelyCause   adb.DiagnosticLikelyCause
		wantAdbRunning    bool
		wantUsbPresent    bool
		wantHeadlineSubstr string
	}{
		{
			name:            "Device ready",
			versionOutput:   "Android Debug Bridge version 1.0.41",
			devicesOutput:   "List of devices attached\nDEV123\tdevice product:XOS model:Infinix_Note_11S device:Infinix-X698 transport_id:1\n",
			usbDevices:      []adb.USBDevice{{DeviceID: `USB\VID_2717&PID_FF40`, Name: "Infinix Note 11S"}},
			wantLikelyCause: adb.CauseReady,
			wantAdbRunning:  true,
			wantUsbPresent:  false, // stops early on ready
			wantHeadlineSubstr: "is connected and authorized",
		},
		{
			name:            "Device unauthorized",
			versionOutput:   "Android Debug Bridge version 1.0.41",
			devicesOutput:   "List of devices attached\nDEV123\tunauthorized transport_id:1\n",
			usbDevices:      []adb.USBDevice{{DeviceID: `USB\VID_2717&PID_FF40`, Name: "Infinix Note 11S"}},
			wantLikelyCause: adb.CauseUnauthorized,
			wantAdbRunning:  true,
			wantUsbPresent:  false, // stops early on unauthorized
			wantHeadlineSubstr: "USB debugging is not authorized",
		},
		{
			name:            "Device offline",
			versionOutput:   "Android Debug Bridge version 1.0.41",
			devicesOutput:   "List of devices attached\nDEV123\toffline transport_id:1\n",
			usbDevices:      []adb.USBDevice{{DeviceID: `USB\VID_2717&PID_FF40`, Name: "Infinix Note 11S"}},
			wantLikelyCause: adb.CauseOffline,
			wantAdbRunning:  true,
			wantUsbPresent:  false, // stops early on offline
			wantHeadlineSubstr: "connection is offline",
		},
		{
			name:            "USB detected but no ADB interface active",
			versionOutput:   "Android Debug Bridge version 1.0.41",
			devicesOutput:   "List of devices attached\n\n",
			usbDevices: []adb.USBDevice{
				{
					DeviceID:               `USB\VID_2717&PID_FF40\0789456`,
					Name:                   "Infinix Note 11S",
					Manufacturer:           "Infinix",
					PNPClass:               "WPD",
					ConfigManagerErrorCode: 0,
				},
			},
			wantLikelyCause:    adb.CauseUsbDetectedNoAdb,
			wantAdbRunning:     true,
			wantUsbPresent:     true,
			wantHeadlineSubstr: "USB device detected by OS, but ADB interface is not active",
		},
		{
			name:               "No USB hardware detected (charge-only or unplugged)",
			versionOutput:      "Android Debug Bridge version 1.0.41",
			devicesOutput:      "List of devices attached\n\n",
			usbDevices:         []adb.USBDevice{},
			wantLikelyCause:    adb.CauseNoUsbHardware,
			wantAdbRunning:     true,
			wantUsbPresent:     false,
			wantHeadlineSubstr: "No Android USB device detected by the operating system",
		},
		{
			name:               "ADB server failed or binary missing",
			versionOutput:      "",
			versionErr:         adb.NewADBError(adb.ErrCodeAdbMissing, "adb not found", nil),
			devicesOutput:      "",
			usbDevices:         []adb.USBDevice{},
			wantLikelyCause:    adb.CauseAdbServerFailed,
			wantAdbRunning:     false,
			wantUsbPresent:     false,
			wantHeadlineSubstr: "ADB server is not responding",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fakeRunner := adb.NewFakeRunner()
			if tc.versionErr != nil {
				fakeRunner.SetResponse("version", "", "command failed", tc.versionErr)
			} else {
				fakeRunner.SetResponse("version", tc.versionOutput, "", nil)
			}
			fakeRunner.SetResponse("devices -l", tc.devicesOutput, "", nil)

			mgr := adb.NewManager(fakeRunner, "/mock/bin/adb", nil)
			fakeDetector := adb.NewFakeUSBDetector(tc.usbDevices...)

			res, err := mgr.DiagnoseConnection(context.Background(), fakeDetector)
			if err != nil {
				t.Fatalf("DiagnoseConnection failed: %v", err)
			}

			if res.AdbServerRunning != tc.wantAdbRunning {
				t.Errorf("AdbServerRunning = %v, want %v", res.AdbServerRunning, tc.wantAdbRunning)
			}
			if res.UsbHardwarePresent != tc.wantUsbPresent {
				t.Errorf("UsbHardwarePresent = %v, want %v", res.UsbHardwarePresent, tc.wantUsbPresent)
			}
			if res.LikelyCause != tc.wantLikelyCause {
				t.Errorf("LikelyCause = %s, want %s", res.LikelyCause, tc.wantLikelyCause)
			}
			if len(tc.wantHeadlineSubstr) > 0 && !containsSubstr(res.Headline, tc.wantHeadlineSubstr) {
				t.Errorf("Headline = %q, want substring %q", res.Headline, tc.wantHeadlineSubstr)
			}
			if len(res.RecommendedSteps) == 0 {
				t.Errorf("expected non-empty RecommendedSteps for cause %s", res.LikelyCause)
			}
		})
	}
}

func containsSubstr(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstr(s, substr)))
}

func findSubstr(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
