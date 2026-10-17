package adb

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/EnoughForYou/Jerion/internal/logx"
)

var (
	reADBVersion   = regexp.MustCompile(`Android Debug Bridge version\s+([0-9.]+)`)
	reHost         = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
	rePairingCode  = regexp.MustCompile(`^[0-9]{6}$`)
)

// Manager manages ADB binary path resolution, version validation, and server lifecycle.
type Manager struct {
	mu         sync.RWMutex
	customPath string
	runner     Runner
	logger     *logx.Logger
}

// NewManager creates an ADB Manager.
func NewManager(runner Runner, customPath string, logger *logx.Logger) *Manager {
	if logger == nil {
		logger = logx.GetDefault()
	}
	return &Manager{
		runner:     runner,
		customPath: customPath,
		logger:     logger,
	}
}

// DefaultAppToolsPath returns the application-managed platform-tools directory.
func DefaultAppToolsPath() (string, error) {
	appData := os.Getenv("LOCALAPPDATA")
	if appData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		appData = filepath.Join(home, ".local", "share")
	}

	execName := "adb"
	if os.PathSeparator == '\\' {
		execName = "adb.exe"
	}
	return filepath.Join(appData, "Jerion", "tools", "platform-tools", execName), nil
}

// ResolveADBPath searches for the ADB executable in priority order:
// 1. Custom path configured in Settings / passed in
// 2. System PATH (exec.LookPath)
// 3. Application-managed tools directory (%LOCALAPPDATA%\Jerion\tools\platform-tools\adb.exe)
func (m *Manager) ResolveADBPath() (string, error) {
	m.mu.RLock()
	custom := m.customPath
	m.mu.RUnlock()

	// 1. Custom path
	if custom != "" {
		if fi, err := os.Stat(custom); err == nil && !fi.IsDir() {
			m.syncRunnerPath(custom)
			return custom, nil
		}
	}

	// 2. System PATH
	if p, err := exec.LookPath("adb"); err == nil {
		m.syncRunnerPath(p)
		return p, nil
	}

	// 3. Application tools directory
	if appTool, err := DefaultAppToolsPath(); err == nil {
		if fi, err := os.Stat(appTool); err == nil && !fi.IsDir() {
			m.syncRunnerPath(appTool)
			return appTool, nil
		}
	}

	return "", NewADBError(ErrCodeAdbMissing, "adb executable not found in custom path, system PATH, or application tools directory", nil)
}

// SetCustomPath updates the custom path override.
func (m *Manager) SetCustomPath(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.customPath = path
	m.syncRunnerPath(path)
}

func (m *Manager) syncRunnerPath(path string) {
	if r, ok := m.runner.(interface{ SetADBPath(string) }); ok {
		r.SetADBPath(path)
	}
}

// CheckVersion executes `adb version` and extracts the semantic version.
func (m *Manager) CheckVersion(ctx context.Context) (string, error) {
	if path, err := m.ResolveADBPath(); err == nil {
		m.syncRunnerPath(path)
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	stdout, stderr, err := m.runner.Run(ctxTimeout, "", "version")
	if err != nil {
		return "", NewADBError(ErrCodeAdbMissing, fmt.Sprintf("failed to invoke adb version: %s", string(stderr)), err)
	}

	outStr := string(stdout)
	match := reADBVersion.FindStringSubmatch(outStr)
	if len(match) >= 2 {
		return match[1], nil
	}

	firstLine := strings.Split(strings.TrimSpace(outStr), "\n")[0]
	if firstLine != "" {
		return firstLine, nil
	}
	return "unknown", nil
}

// StartServer starts the ADB server daemon.
func (m *Manager) StartServer(ctx context.Context) error {
	if path, err := m.ResolveADBPath(); err == nil {
		m.syncRunnerPath(path)
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, stderr, err := m.runner.Run(ctxTimeout, "", "start-server")
	if err != nil {
		return NewADBError(ErrCodeCommandFailed, fmt.Sprintf("failed to start adb server: %s", string(stderr)), err)
	}
	return nil
}

// KillServer terminates any running ADB server daemon.
func (m *Manager) KillServer(ctx context.Context) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, stderr, err := m.runner.Run(ctxTimeout, "", "kill-server")
	if err != nil {
		return NewADBError(ErrCodeCommandFailed, fmt.Sprintf("failed to kill adb server: %s", string(stderr)), err)
	}
	return nil
}

// PairWireless executes `adb pair <host>:<pairingPort> <pairingCode>` to pair with an Android 11+ (API 30+) device.
// The pairingCode is validated to be exactly 6 digits and is never exposed in error messages or debug logs.
func (m *Manager) PairWireless(ctx context.Context, host string, pairingPort int, pairingCode string) error {
	if host == "" || !reHost.MatchString(host) {
		return NewADBError(ErrCodeValidationFailed, fmt.Sprintf("invalid host format: %q", host), nil)
	}
	if pairingPort < 1 || pairingPort > 65535 {
		return NewADBError(ErrCodeValidationFailed, fmt.Sprintf("invalid pairing port: %d (must be 1-65535)", pairingPort), nil)
	}
	if !rePairingCode.MatchString(pairingCode) {
		return NewADBError(ErrCodeValidationFailed, "pairing code must be exactly 6 digits", nil)
	}

	if path, err := m.ResolveADBPath(); err == nil {
		m.syncRunnerPath(path)
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	endpoint := fmt.Sprintf("%s:%d", host, pairingPort)
	stdout, stderr, err := m.runner.Run(ctxTimeout, "", "pair", endpoint, pairingCode)
	outStr := strings.TrimSpace(string(stdout) + " " + string(stderr))

	if err != nil {
		return NewADBError(ErrCodeCommandFailed, fmt.Sprintf("wireless pairing failed: %s", outStr), err)
	}

	lower := strings.ToLower(outStr)
	if strings.Contains(lower, "failed") || strings.Contains(lower, "error") || strings.Contains(lower, "unable to") {
		return NewADBError(ErrCodeCommandFailed, fmt.Sprintf("wireless pairing rejected: %s", outStr), nil)
	}

	return nil
}

// ConnectWireless executes `adb connect <host>:<connectionPort>` to establish an active ADB session.
func (m *Manager) ConnectWireless(ctx context.Context, host string, connectionPort int) (string, error) {
	if host == "" || !reHost.MatchString(host) {
		return "", NewADBError(ErrCodeValidationFailed, fmt.Sprintf("invalid host format: %q", host), nil)
	}
	if connectionPort < 1 || connectionPort > 65535 {
		return "", NewADBError(ErrCodeValidationFailed, fmt.Sprintf("invalid connection port: %d (must be 1-65535)", connectionPort), nil)
	}

	if path, err := m.ResolveADBPath(); err == nil {
		m.syncRunnerPath(path)
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	endpoint := fmt.Sprintf("%s:%d", host, connectionPort)
	stdout, stderr, err := m.runner.Run(ctxTimeout, "", "connect", endpoint)
	outStr := strings.TrimSpace(string(stdout) + " " + string(stderr))

	if err != nil {
		return "", NewADBError(ErrCodeCommandFailed, fmt.Sprintf("failed to connect wirelessly: %s", outStr), err)
	}

	lower := strings.ToLower(outStr)
	if strings.Contains(lower, "cannot connect") || strings.Contains(lower, "failed") || strings.Contains(lower, "unable to connect") {
		return "", NewADBError(ErrCodeCommandFailed, fmt.Sprintf("wireless connection rejected: %s", outStr), nil)
	}

	return endpoint, nil
}

// DisconnectWireless executes `adb disconnect <host>:<connectionPort>`.
func (m *Manager) DisconnectWireless(ctx context.Context, host string, connectionPort int) error {
	if host == "" || !reHost.MatchString(host) {
		return NewADBError(ErrCodeValidationFailed, fmt.Sprintf("invalid host format: %q", host), nil)
	}
	if connectionPort < 1 || connectionPort > 65535 {
		return NewADBError(ErrCodeValidationFailed, fmt.Sprintf("invalid connection port: %d (must be 1-65535)", connectionPort), nil)
	}

	if path, err := m.ResolveADBPath(); err == nil {
		m.syncRunnerPath(path)
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	endpoint := fmt.Sprintf("%s:%d", host, connectionPort)
	_, stderr, err := m.runner.Run(ctxTimeout, "", "disconnect", endpoint)
	if err != nil {
		return NewADBError(ErrCodeCommandFailed, fmt.Sprintf("failed to disconnect %s: %s", endpoint, string(stderr)), err)
	}
	return nil
}

// SoftReconnect attempts to reset the ADB connection socket without restarting the daemon.
// If serial is provided, it executes `adb -s <serial> reconnect`. Otherwise it executes `adb reconnect`.
func (m *Manager) SoftReconnect(ctx context.Context, serial string) error {
	if err := ValidateSerial(serial); err != nil {
		return err
	}

	if path, err := m.ResolveADBPath(); err == nil {
		m.syncRunnerPath(path)
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, stderr, err := m.runner.Run(ctxTimeout, serial, "reconnect")
	if err != nil {
		return NewADBError(ErrCodeCommandFailed, fmt.Sprintf("soft reconnect failed: %s", string(stderr)), err)
	}
	return nil
}
