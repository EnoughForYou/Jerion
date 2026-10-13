package adb

import (
	"regexp"
	"strconv"
	"strings"
)

// DeviceState represents ADB device connection state.
type DeviceState string

const (
	StateReady        DeviceState = "ready"
	StateUnauthorized DeviceState = "unauthorized"
	StateOffline      DeviceState = "offline"
	StateNoDevice     DeviceState = "no_device"
	StateAdbMissing   DeviceState = "adb_missing"
)

// DeviceInfo holds parsed properties of an attached device.
type DeviceInfo struct {
	Serial       string      `json:"serial"`
	State        DeviceState `json:"state"`
	Product      string      `json:"product,omitempty"`
	Model        string      `json:"model,omitempty"`
	DeviceName   string      `json:"device,omitempty"`
	TransportID  string      `json:"transportId,omitempty"`
	RawState     string      `json:"rawState"`
}

// BatteryInfo represents parsed battery telemetry.
type BatteryInfo struct {
	Level       int     `json:"level"`
	Scale       int     `json:"scale"`
	Percent     int     `json:"percent"`
	Status      string  `json:"status"`
	Health      string  `json:"health"`
	Temperature float64 `json:"temperature"`
	IsCharging  bool    `json:"isCharging"`
	VoltageMV   int     `json:"voltageMv"`
}

// StorageInfo represents storage metrics for /data partition.
type StorageInfo struct {
	TotalBytes int64 `json:"totalBytes"`
	UsedBytes  int64 `json:"usedBytes"`
	FreeBytes  int64 `json:"freeBytes"`
	UsePercent int   `json:"usePercent"`
}

// WindowFocusInfo captures current foreground window and activity.
type WindowFocusInfo struct {
	Package  string `json:"package"`
	Activity string `json:"activity"`
	RawFocus string `json:"rawFocus"`
}

var (
	reGetprop     = regexp.MustCompile(`\[([^\]]+)\]:\s*\[([^\]]*)\]`)
	reWindowFocus = regexp.MustCompile(`(?:mCurrentFocus|mFocusedApp).*?([a-zA-Z0-9_.]+)/([a-zA-Z0-9_.]+)`)
)

// ParseDevices parses stdout from `adb devices -l`.
func ParseDevices(output string) []DeviceInfo {
	var devices []DeviceInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "List of devices") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		fields := strings.Fields(trimmed)
		if len(fields) < 2 {
			continue
		}

		serial := fields[0]
		rawState := fields[1]

		var state DeviceState
		switch rawState {
		case "device":
			state = StateReady
		case "unauthorized":
			state = StateUnauthorized
		case "offline":
			state = StateOffline
		default:
			state = DeviceState(rawState)
		}

		info := DeviceInfo{
			Serial:   serial,
			State:    state,
			RawState: rawState,
		}

		// Parse key:value fields like model:Infinix_X698
		for _, f := range fields[2:] {
			parts := strings.SplitN(f, ":", 2)
			if len(parts) != 2 {
				continue
			}
			k, v := parts[0], parts[1]
			switch k {
			case "product":
				info.Product = v
			case "model":
				info.Model = v
			case "device":
				info.DeviceName = v
			case "transport_id":
				info.TransportID = v
			}
		}

		devices = append(devices, info)
	}

	return devices
}

// ParseGetprop parses stdout from `getprop`.
func ParseGetprop(output string) map[string]string {
	props := make(map[string]string)
	matches := reGetprop.FindAllStringSubmatch(output, -1)
	for _, m := range matches {
		if len(m) == 3 {
			props[m[1]] = m[2]
		}
	}
	return props
}

// ParseBattery defensively parses `dumpsys battery` output.
func ParseBattery(output string) BatteryInfo {
	info := BatteryInfo{
		Scale: 100,
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])

		switch key {
		case "level":
			if n, err := strconv.Atoi(val); err == nil {
				info.Level = n
			}
		case "scale":
			if n, err := strconv.Atoi(val); err == nil && n > 0 {
				info.Scale = n
			}
		case "voltage":
			if n, err := strconv.Atoi(val); err == nil {
				info.VoltageMV = n
			}
		case "temperature":
			if n, err := strconv.Atoi(val); err == nil {
				info.Temperature = float64(n) / 10.0 // tenths of degree Celsius
			}
		case "status":
			switch val {
			case "2":
				info.Status = "Charging"
				info.IsCharging = true
			case "3":
				info.Status = "Discharging"
			case "4":
				info.Status = "Not charging"
			case "5":
				info.Status = "Full"
			default:
				info.Status = val
			}
		case "health":
			switch val {
			case "2":
				info.Health = "Good"
			case "3":
				info.Health = "Overheat"
			case "4":
				info.Health = "Dead"
			case "5":
				info.Health = "Over voltage"
			case "6":
				info.Health = "Unspecified failure"
			case "7":
				info.Health = "Cold"
			default:
				info.Health = val
			}
		case "ac powered", "usb powered", "wireless powered":
			if strings.EqualFold(val, "true") {
				info.IsCharging = true
			}
		}
	}

	if info.Scale > 0 {
		info.Percent = (info.Level * 100) / info.Scale
	} else {
		info.Percent = info.Level
	}

	return info
}

// ParseStorage defensively parses `df /data` or `df -k /data` output.
func ParseStorage(output string) StorageInfo {
	var info StorageInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "Filesystem") {
			continue
		}

		fields := strings.Fields(trimmed)
		// Expected: Filesystem 1K-blocks Used Available Use% Mounted
		// E.g.: /dev/block/dm-5 116245084 45689124 70555960 40% /data
		if len(fields) >= 5 {
			var total, used, free int64
			var pct int

			// Attempt 1K-blocks parse
			if t, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
				total = t * 1024
			}
			if u, err := strconv.ParseInt(fields[2], 10, 64); err == nil {
				used = u * 1024
			}
			if a, err := strconv.ParseInt(fields[3], 10, 64); err == nil {
				free = a * 1024
			}

			// Parse Use%
			pctStr := strings.TrimSuffix(fields[4], "%")
			if p, err := strconv.Atoi(pctStr); err == nil {
				pct = p
			}

			if total > 0 {
				info.TotalBytes = total
				info.UsedBytes = used
				info.FreeBytes = free
				info.UsePercent = pct
				return info
			}
		}
	}

	return info
}

// ParseWindowFocus extracts active package and activity from `dumpsys window`.
func ParseWindowFocus(output string) WindowFocusInfo {
	var info WindowFocusInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "mCurrentFocus") || strings.Contains(trimmed, "mFocusedApp") {
			info.RawFocus = trimmed
			m := reWindowFocus.FindStringSubmatch(trimmed)
			if len(m) == 3 {
				info.Package = m[1]
				info.Activity = m[2]
				return info
			}
		}
	}

	return info
}
