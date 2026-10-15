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

var reADBVersion = regexp.MustCompile(`Android Debug Bridge version\s+([0-9.]+)`)

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
