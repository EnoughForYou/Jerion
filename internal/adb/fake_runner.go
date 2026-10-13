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

	// 3. Fallback
	if f.defaultSet {
		return f.defaultResp.stdout, f.defaultResp.stderr, f.defaultResp.err
	}

	return []byte(""), []byte(""), nil
}
