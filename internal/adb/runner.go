package adb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/EnoughForYou/Jerion/internal/logx"
)

var (
	rePackageName    = regexp.MustCompile(`^[a-zA-Z0-9_.]+$`)
	rePermissionName = regexp.MustCompile(`^[a-zA-Z0-9_.]+$`)
	reSerial         = regexp.MustCompile(`^[a-zA-Z0-9._:-]+$`)
)

// ValidatePackageName ensures package identifier contains only allowed characters.
func ValidatePackageName(pkg string) error {
	if !rePackageName.MatchString(pkg) {
		return NewADBError(ErrCodeValidationFailed, fmt.Sprintf("invalid package name format: %q", pkg), nil)
	}
	return nil
}

// ValidatePermissionName ensures permission identifier contains only allowed characters.
func ValidatePermissionName(perm string) error {
	if !rePermissionName.MatchString(perm) {
		return NewADBError(ErrCodeValidationFailed, fmt.Sprintf("invalid permission name format: %q", perm), nil)
	}
	return nil
}

// ValidateSerial ensures device serial or host:port contains only allowed characters.
func ValidateSerial(serial string) error {
	if serial == "" {
		return nil
	}
	if !reSerial.MatchString(serial) {
		return NewADBError(ErrCodeValidationFailed, fmt.Sprintf("invalid serial format: %q", serial), nil)
	}
	return nil
}

// sanitizeArgsForLog creates a copy of command arguments with sensitive values like pairing code redacted.
func sanitizeArgsForLog(args []string) []string {
	if len(args) == 0 {
		return args
	}
	clean := make([]string, len(args))
	copy(clean, args)
	for i := 0; i < len(clean); i++ {
		if clean[i] == "pair" && i+2 < len(clean) {
			clean[i+2] = "[REDACTED]"
		}
	}
	return clean
}

// Runner abstracts executing ADB commands on real devices or simulated environments.
type Runner interface {
	Run(ctx context.Context, serial string, args ...string) (stdout []byte, stderr []byte, err error)
}

// RealRunner executes native ADB binary commands via os/exec.
type RealRunner struct {
	adbPath      string
	logger       *logx.Logger
	mu           sync.Mutex
	deviceQueues map[string]*sync.Mutex
}

// NewRealRunner constructs an operational RealRunner.
func NewRealRunner(adbPath string, logger *logx.Logger) *RealRunner {
	if logger == nil {
		logger = logx.GetDefault()
	}
	return &RealRunner{
		adbPath:      adbPath,
		logger:       logger,
		deviceQueues: make(map[string]*sync.Mutex),
	}
}

// SetADBPath updates the active executable path.
func (r *RealRunner) SetADBPath(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adbPath = path
}

// ADBPath returns the currently configured binary path.
func (r *RealRunner) ADBPath() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.adbPath
}

func (r *RealRunner) getDeviceLock(serial string) *sync.Mutex {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.deviceQueues[serial]; !ok {
		r.deviceQueues[serial] = &sync.Mutex{}
	}
	return r.deviceQueues[serial]
}

// Run executes the command using the configured ADB binary.
func (r *RealRunner) Run(ctx context.Context, serial string, args ...string) ([]byte, []byte, error) {
	activePath := r.ADBPath()
	if activePath == "" || activePath == "adb" {
		if _, err := exec.LookPath(activePath); err != nil {
			if appTool, err := DefaultAppToolsPath(); err == nil {
				if fi, err := os.Stat(appTool); err == nil && !fi.IsDir() {
					r.SetADBPath(appTool)
					activePath = appTool
				}
			}
		}
	}
	if activePath == "" {
		return nil, nil, NewADBError(ErrCodeAdbMissing, "adb binary path is not configured", nil)
	}

	// Prepare arguments
	fullArgs := make([]string, 0, len(args)+2)
	if serial != "" {
		fullArgs = append(fullArgs, "-s", serial)
	}
	fullArgs = append(fullArgs, args...)

	// Acquire per-device queue lock to avoid overwhelming ADB daemon
	devLock := r.getDeviceLock(serial)
	devLock.Lock()
	defer devLock.Unlock()

	startTime := time.Now()
	cmd := exec.CommandContext(ctx, activePath, fullArgs...)
	setHideWindow(cmd)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	duration := time.Since(startTime)

	stdout := stdoutBuf.Bytes()
	stderr := stderrBuf.Bytes()
	stderrStr := strings.ToLower(string(stderr))

	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	// Audit via structured logger with sensitive arguments redacted
	logArgs := sanitizeArgsForLog(fullArgs)
	r.logger.DebugADB(serial, r.adbPath, logArgs, duration, exitCode, err)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(ctx.Err(), context.Canceled) {
			return stdout, stderr, NewADBError(ErrCodeTimeout, "adb command execution timed out", ctx.Err())
		}
		if strings.Contains(stderrStr, "device unauthorized") || strings.Contains(stderrStr, "unauthorized") {
			return stdout, stderr, NewADBError(ErrCodeDeviceUnauthorized, "device is unauthorized; prompt must be accepted on phone", err)
		}
		if strings.Contains(stderrStr, "device offline") || strings.Contains(stderrStr, "device not found") {
			return stdout, stderr, NewADBError(ErrCodeDeviceOffline, "device is offline or disconnected", err)
		}
		if strings.Contains(stderrStr, "permission denied") {
			return stdout, stderr, NewADBError(ErrCodePermissionDenied, "permission denied on android device", err)
		}

		return stdout, stderr, NewADBError(ErrCodeCommandFailed, fmt.Sprintf("adb command exited with code %d", exitCode), err)
	}

	return stdout, stderr, nil
}
