package adb

import (
	"context"
	"strings"
	"sync"
)

// CallRecord captures an invocation of Runner.Run for assertions in tests.
type CallRecord struct {
	Serial string
	Args   []string
}

// FullCommand returns the joined arguments.
func (c CallRecord) FullCommand() string {
	return strings.Join(c.Args, " ")
}

type cannedResponse struct {
	stdout []byte
	stderr []byte
	err    error
}

type handlerFunc func(serial string, args []string) ([]byte, []byte, error)

type handlerRule struct {
	predicate func(serial string, args []string) bool
	handler   handlerFunc
}

// FakeRunner provides a thread-safe Runner mock for unit testing without physical devices.
type FakeRunner struct {
	mu           sync.Mutex
	adbPath      string
	calls        []CallRecord
	responses    map[string]cannedResponse
	handlers     []handlerRule
	defaultResp  cannedResponse
	defaultSet   bool
}

// NewFakeRunner creates a new FakeRunner.
func NewFakeRunner() *FakeRunner {
	return &FakeRunner{
		responses: make(map[string]cannedResponse),
	}
}

// SetADBPath records the active executable path.
func (f *FakeRunner) SetADBPath(path string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.adbPath = path
}

// ADBPath returns the active executable path.
func (f *FakeRunner) ADBPath() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.adbPath
}

// SetResponse registers canned output when args joined contain commandSubstr.
func (f *FakeRunner) SetResponse(commandSubstr string, stdout, stderr string, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.responses[commandSubstr] = cannedResponse{
		stdout: []byte(stdout),
		stderr: []byte(stderr),
		err:    err,
	}
}

// SetHandler registers a dynamic handler rule evaluated in insertion order.
func (f *FakeRunner) SetHandler(predicate func(serial string, args []string) bool, handler handlerFunc) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.handlers = append(f.handlers, handlerRule{
		predicate: predicate,
		handler:   handler,
	})
}

// SetDefaultResponse configures fallback output when no pattern or handler matches.
func (f *FakeRunner) SetDefaultResponse(stdout, stderr string, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.defaultResp = cannedResponse{
		stdout: []byte(stdout),
		stderr: []byte(stderr),
		err:    err,
	}
	f.defaultSet = true
}

// Calls returns a copy of all recorded invocations.
func (f *FakeRunner) Calls() []CallRecord {
	f.mu.Lock()
	defer f.mu.Unlock()
	res := make([]CallRecord, len(f.calls))
	copy(res, f.calls)
	return res
}

// ResetCalls clears recorded calls.
func (f *FakeRunner) ResetCalls() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = nil
}

// Run executes the mock logic.
func (f *FakeRunner) Run(ctx context.Context, serial string, args ...string) ([]byte, []byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls = append(f.calls, CallRecord{
		Serial: serial,
		Args:   args,
	})

	fullCmd := strings.Join(args, " ")

	// 1. Check dynamic handlers
	for _, rule := range f.handlers {
		if rule.predicate(serial, args) {
			return rule.handler(serial, args)
		}
	}

	// 2. Check substring canned responses
	for substr, resp := range f.responses {
		if strings.Contains(fullCmd, substr) {
			return resp.stdout, resp.stderr, resp.err
		}
	}

	return []byte(""), []byte(""), nil
}

// FakeUSBDetector provides controllable USB hardware device detection for tests.
type FakeUSBDetector struct {
	mu      sync.Mutex
	devices []USBDevice
	err     error
}

// NewFakeUSBDetector creates a test detector.
func NewFakeUSBDetector(devices ...USBDevice) *FakeUSBDetector {
	return &FakeUSBDetector{
		devices: devices,
	}
}

// SetDevices updates the simulated connected USB hardware.
func (f *FakeUSBDetector) SetDevices(devices []USBDevice) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.devices = devices
}

// SetError updates the simulated OS error.
func (f *FakeUSBDetector) SetError(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.err = err
}

// DetectUSBDevices returns the simulated devices.
func (f *FakeUSBDetector) DetectUSBDevices(ctx context.Context) ([]USBDevice, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return nil, f.err
	}
	res := make([]USBDevice, len(f.devices))
	copy(res, f.devices)
	return res, nil
}
