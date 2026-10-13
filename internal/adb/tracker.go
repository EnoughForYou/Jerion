package adb

import (
	"context"
	"sync"
	"time"

	"github.com/EnoughForYou/Jerion/internal/logx"
)

// DeviceChangeListener receives updates when attached devices change.
type DeviceChangeListener func(devices []DeviceInfo, active *DeviceInfo)

// Tracker monitors connected Android devices and maintains the active device state.
type Tracker struct {
	runner       Runner
	logger       *logx.Logger
	mu           sync.RWMutex
	devices      []DeviceInfo
	activeSerial string
	listeners    []DeviceChangeListener
	cancelPoll   context.CancelFunc
}

// NewTracker creates a new Device Tracker.
func NewTracker(runner Runner, logger *logx.Logger) *Tracker {
	if logger == nil {
		logger = logx.GetDefault()
	}
	return &Tracker{
		runner: runner,
		logger: logger,
	}
}

// AddListener registers a callback invoked when device topology changes.
func (t *Tracker) AddListener(listener DeviceChangeListener) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.listeners = append(t.listeners, listener)
}

// SetActiveSerial sets or overrides the selected target device serial.
func (t *Tracker) SetActiveSerial(serial string) {
	t.mu.Lock()
	t.activeSerial = serial
	devs := t.devices
	active := t.findActiveLocked()
	listeners := t.listeners
	t.mu.Unlock()

	for _, l := range listeners {
		l(devs, active)
	}
}

func (t *Tracker) findActiveLocked() *DeviceInfo {
	if len(t.devices) == 0 {
		return nil
	}
	if t.activeSerial != "" {
		for i := range t.devices {
			if t.devices[i].Serial == t.activeSerial {
				return &t.devices[i]
			}
		}
	}
	// Fallback to first ready device, or first device if none ready
	for i := range t.devices {
		if t.devices[i].State == StateReady {
			return &t.devices[i]
		}
	}
	return &t.devices[0]
}

// ActiveDevice returns the currently focused device or nil.
func (t *Tracker) ActiveDevice() *DeviceInfo {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.findActiveLocked()
}

// Devices returns a snapshot of all discovered devices.
func (t *Tracker) Devices() []DeviceInfo {
	t.mu.RLock()
	defer t.mu.RUnlock()
	res := make([]DeviceInfo, len(t.devices))
	copy(res, t.devices)
	return res
}

// PollOnce executes `devices -l` once and notifies listeners if state changed.
func (t *Tracker) PollOnce(ctx context.Context) ([]DeviceInfo, error) {
	stdout, _, err := t.runner.Run(ctx, "", "devices", "-l")
	if err != nil {
		return nil, err
	}

	current := ParseDevices(string(stdout))

	t.mu.Lock()
	changed := !devicesEqual(t.devices, current)
	t.devices = current
	active := t.findActiveLocked()
	listeners := t.listeners
	t.mu.Unlock()

	if changed {
		for _, l := range listeners {
			l(current, active)
		}
	}

	return current, nil
}

// Start begins periodic device tracking in a background goroutine.
func (t *Tracker) Start(ctx context.Context, interval time.Duration) {
	t.mu.Lock()
	if t.cancelPoll != nil {
		t.mu.Unlock()
		return
	}
	pollCtx, cancel := context.WithCancel(ctx)
	t.cancelPoll = cancel
	t.mu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Initial immediate poll
		_, _ = t.PollOnce(pollCtx)

		for {
			select {
			case <-pollCtx.Done():
				return
			case <-ticker.C:
				_, _ = t.PollOnce(pollCtx)
			}
		}
	}()
}

// Stop terminates background device polling.
func (t *Tracker) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.cancelPoll != nil {
		t.cancelPoll()
		t.cancelPoll = nil
	}
}

func devicesEqual(a, b []DeviceInfo) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Serial != b[i].Serial || a[i].State != b[i].State || a[i].Model != b[i].Model {
			return false
		}
	}
	return true
}
