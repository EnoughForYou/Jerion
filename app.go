package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/EnoughForYou/Jerion/internal/adb"
	"github.com/EnoughForYou/Jerion/internal/apk"
	"github.com/EnoughForYou/Jerion/internal/audit"
	"github.com/EnoughForYou/Jerion/internal/contract"
	"github.com/EnoughForYou/Jerion/internal/engine"
	"github.com/EnoughForYou/Jerion/internal/intel"
	"github.com/EnoughForYou/Jerion/internal/inventory"
	"github.com/EnoughForYou/Jerion/internal/logx"
	"github.com/EnoughForYou/Jerion/internal/monitor"
	"github.com/EnoughForYou/Jerion/internal/preflight"
	"github.com/EnoughForYou/Jerion/internal/remediate"
	"github.com/EnoughForYou/Jerion/internal/report"
	"github.com/EnoughForYou/Jerion/internal/rules"
	"github.com/EnoughForYou/Jerion/internal/secrets"
	"github.com/EnoughForYou/Jerion/internal/store"
	"github.com/EnoughForYou/Jerion/internal/version"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func wireError(code, message, remedy string) error {
	return fmt.Errorf(`JERION_ERR:{"code":"%s","message":"%s","remedy":"%s"}`, code, message, remedy)
}

func wrapADBError(err error) error {
	if err == nil {
		return nil
	}
	var adbErr *adb.ADBError
	if errors.As(err, &adbErr) {
		switch adbErr.Code {
		case adb.ErrCodeAdbMissing:
			return wireError("AdbMissing", adbErr.Message, "Install platform-tools or configure custom ADB path in Settings")
		case adb.ErrCodeDeviceUnauthorized:
			return wireError("DeviceUnauthorized", adbErr.Message, "Check phone screen and authorize computer RSA key")
		case adb.ErrCodeDeviceOffline:
			return wireError("DeviceOffline", adbErr.Message, "Reconnect USB cable or toggle USB debugging")
		case adb.ErrCodeTimeout:
			return wireError("Timeout", adbErr.Message, "Command took too long to respond; verify device responsiveness")
		case adb.ErrCodePermissionDenied:
			return wireError("PermissionDenied", adbErr.Message, "Operation requires elevated Android permissions")
		case adb.ErrCodeValidationFailed:
			return wireError("ValidationFailed", adbErr.Message, "Verify input parameters")
		case adb.ErrCodeCommandFailed:
			return wireError("CommandFailed", adbErr.Message, "Check ADB command output or device settings")
		default:
			return wireError("InternalError", adbErr.Message, "Check ADB and device state")
		}
	}
	return wireError("InternalError", err.Error(), "Check application logs")
}

var rePermissionSnippet = regexp.MustCompile(`[a-zA-Z0-9_.]+\.permission\.[a-zA-Z0-9_]+`)

func extractCleanPermission(raw string) string {
	raw = strings.TrimSpace(raw)
	if m := rePermissionSnippet.FindString(raw); m != "" {
		return m
	}
	for _, p := range []string{"Declared/Requested:", "Requested:", "Granted:", "Declared:"} {
		if strings.HasPrefix(raw, p) {
			raw = strings.TrimSpace(strings.TrimPrefix(raw, p))
		}
	}
	if adb.ValidatePermissionName(raw) == nil {
		return raw
	}
	return ""
}

func containsString(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// App encapsulates all backend services exposed to Wails v2.
type App struct {
	ctx           context.Context
	mu            sync.RWMutex
	runner        adb.Runner
	store         *store.Store
	logger        *logx.Logger
	secrets       *secrets.Manager
	adbMgr        *adb.Manager
	tracker       *adb.Tracker
	preflight     *preflight.Engine
	inventory     *inventory.Service
	puller        *inventory.Puller
	scanner       *engine.Scanner
	monitor       *monitor.Manager
	auditSvc      *audit.Service
	remediate     *remediate.Manager
	intelAgg      *intel.Aggregator
	usbDetector   adb.USBDetector
	workDir       string
	activeSerial  string
	latestAudit   []contract.AuditItem
	activeScans   map[string]context.CancelFunc
	cachedResults map[string]*contract.ScanResult
	cachedActions map[string]remediate.Action
}

// NewApp initializes production backend services.
func NewApp() *App {
	logger := logx.GetDefault()
	runner := adb.NewRealRunner("adb", logger)

	dbPath, err := store.DefaultDBPath()
	if err != nil {
		dbPath = filepath.Join(os.TempDir(), "jerion.db")
	}
	st, err := store.Open(dbPath)
	if err != nil {
		st, _ = store.OpenMemory()
	}

	return NewAppWithDeps(runner, st, logger, "")
}

// NewAppWithDeps constructs an App instance with injected dependencies for testing.
func NewAppWithDeps(runner adb.Runner, st *store.Store, logger *logx.Logger, workDir string) *App {
	if logger == nil {
		logger = logx.GetDefault()
	}
	if workDir == "" {
		appData := os.Getenv("LOCALAPPDATA")
		if appData == "" {
			appData = os.TempDir()
		}
		workDir = filepath.Join(appData, "Jerion", "work")
	}
	_ = os.MkdirAll(workDir, 0750)

	secMgr := secrets.NewManager(secrets.NewOSKeyring())
	adbMgr := adb.NewManager(runner, "", logger)
	if savedPath, err := st.GetSetting("adbPath"); err == nil && savedPath != "" {
		adbMgr.SetCustomPath(savedPath)
	} else if resolvedPath, err := adbMgr.ResolveADBPath(); err == nil && resolvedPath != "" {
		if psr, ok := runner.(interface{ SetADBPath(string) }); ok {
			psr.SetADBPath(resolvedPath)
		}
	}
	tracker := adb.NewTracker(runner, logger)
	prefEngine := preflight.NewEngine(adbMgr, secMgr, st, workDir)

	invSvc := inventory.NewService(runner, logger)
	puller := inventory.NewPuller(runner, logger)
	inspector := apk.NewInspector()
	intelAgg := intel.NewAggregator(st, secMgr)

	// Configure local ClamAV client if enabled in persisted settings
	if allSettings, err := st.GetAllSettings(); err == nil {
		if allSettings["clamav_enabled"] == "true" {
			clamHost := intel.DefaultClamAVAddress
			if h := allSettings["clamav_host"]; h != "" {
				if valid, valErr := intel.ValidateClamAVAddress(h); valErr == nil {
					clamHost = valid
				}
			}
			intelAgg.SetClamAVClient(intel.NewClamAVClient(clamHost))
		}
	}

	sc, err := engine.NewScanner(invSvc, puller, inspector, intelAgg, st, workDir, logger)
	if err != nil {
		logger.Error("failed to create scanner", "err", err)
	}

	monMgr := monitor.NewManager(runner, logger)
	auditSvc := audit.NewService(runner, logger)
	remMgr := remediate.NewManager(runner, st, logger, filepath.Join(workDir, "backups"))

	app := &App{
		runner:        runner,
		store:         st,
		logger:        logger,
		secrets:       secMgr,
		adbMgr:        adbMgr,
		tracker:       tracker,
		preflight:     prefEngine,
		inventory:     invSvc,
		puller:        puller,
		scanner:       sc,
		monitor:       monMgr,
		auditSvc:      auditSvc,
		remediate:     remMgr,
		intelAgg:      intelAgg,
		usbDetector:   adb.DefaultUSBDetector(),
		workDir:       workDir,
		activeScans:   make(map[string]context.CancelFunc),
		cachedResults: make(map[string]*contract.ScanResult),
		cachedActions: make(map[string]remediate.Action),
	}

	return app
}

// SetUSBDetector overrides the default USB hardware detector (useful for testing).
func (a *App) SetUSBDetector(detector adb.USBDetector) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.usbDetector = detector
}

// startup is called by Wails when the GUI window opens.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Device changed event emitter
	a.tracker.AddListener(func(devices []adb.DeviceInfo, active *adb.DeviceInfo) {
		a.mu.Lock()
		activeSerial := a.activeSerial
		if active == nil || (activeSerial != "" && active.Serial != activeSerial) {
			// Cancel active scans for disconnected or superseded device
			for scanID, cancel := range a.activeScans {
				a.logger.Warn("device disconnected during active scan; cancelling cleanly", "scanId", scanID, "serial", activeSerial)
				cancel()
				delete(a.activeScans, scanID)
			}
			// Invalidate cached remediation actions and results on device change or disconnect
			a.cachedActions = make(map[string]remediate.Action)
			a.cachedResults = make(map[string]*contract.ScanResult)

			if active == nil {
				a.activeSerial = ""
			} else {
				a.activeSerial = active.Serial
			}
		}
		a.mu.Unlock()

		if a.ctx != nil {
			var devDto *contract.Device
			if active != nil {
				devDto = a.convertDevice(*active)
			}
			wailsRuntime.EventsEmit(a.ctx, "device:changed", devDto)
		}
	})

	// Monitor event listeners
	a.monitor.OnEvent(func(evt monitor.MonitorEvent) {
		if a.ctx != nil {
			wailsRuntime.EventsEmit(a.ctx, "monitor:event", evt)
		}
	})

	a.monitor.OnCandidate(func(cand monitor.MonitorCandidate) {
		if a.ctx != nil {
			wailsRuntime.EventsEmit(a.ctx, "monitor:candidate", cand)
		}
	})

	// Start background device tracker
	a.tracker.Start(ctx, 2*time.Second)
}

// --- 1. Preflight Service ---

func (a *App) PreflightRun() ([]contract.PreflightCheck, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	checks, _ := a.preflight.Run(ctx)
	var dtos []contract.PreflightCheck
	for _, c := range checks {
		dto := a.convertPreflightCheck(c)
		dtos = append(dtos, dto)
		if a.ctx != nil {
			wailsRuntime.EventsEmit(a.ctx, "preflight:update", dto)
		}
	}
	return dtos, nil
}

func (a *App) PreflightStatus() ([]contract.PreflightCheck, error) {
	return a.PreflightRun()
}

func (a *App) PreflightDownloadAdb() error {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	downloader := preflight.NewDownloader()
	expectedSHA := preflight.DefaultPlatformToolsSHA256()
	destPath, err := downloader.DownloadAndExtract(ctx, expectedSHA, func(downloadedBytes, totalBytes int64) {
		if totalBytes > 0 && a.ctx != nil {
			pct := float64(downloadedBytes) / float64(totalBytes) * 100.0
			remedy := "Downloading official Google platform-tools..."
			wailsRuntime.EventsEmit(a.ctx, "preflight:update", contract.PreflightCheck{
				ID:               "adb_binary",
				Title:            "ADB Binary",
				Status:           "downloading",
				Detail:           fmt.Sprintf("Downloading... %.0f%%", pct),
				Remedy:           &remedy,
				DownloadProgress: &pct,
			})
		}
	})
	if err != nil {
		return wireError("InternalError", err.Error(), "Ensure network connection and retry")
	}

	a.adbMgr.SetCustomPath(destPath)
	_ = a.store.SetSetting("adbPath", destPath)
	if psr, ok := a.runner.(interface{ SetADBPath(string) }); ok {
		psr.SetADBPath(destPath)
	}

	return nil
}

func (a *App) PreflightSelectAdbPath(path string) error {
	if strings.TrimSpace(path) == "" {
		if a.ctx == nil {
			return wireError("InternalError", "Window context not ready", "Wait for application window to appear")
		}
		selected, err := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
			Title: "Select ADB Executable (adb.exe)",
			Filters: []wailsRuntime.FileFilter{
				{DisplayName: "ADB Executable (adb.exe; adb)", Pattern: "adb.exe;adb"},
				{DisplayName: "All Files (*.*)", Pattern: "*.*"},
			},
		})
		if err != nil {
			return wireError("InternalError", err.Error(), "Browse and select valid adb.exe executable")
		}
		if selected == "" {
			return nil // User cancelled file picker dialog
		}
		path = selected
	}

	if _, err := os.Stat(path); err != nil {
		return wireError("AdbMissing", "Specified file does not exist", "Browse and select valid adb.exe executable")
	}
	a.adbMgr.SetCustomPath(path)
	_ = a.store.SetSetting("adbPath", path)
	if psr, ok := a.runner.(interface{ SetADBPath(string) }); ok {
		psr.SetADBPath(path)
	}
	return nil
}

func (a *App) PreflightRetry() ([]contract.PreflightCheck, error) {
	return a.PreflightRun()
}

func (a *App) convertPreflightCheck(c preflight.PreflightCheck) contract.PreflightCheck {
	idMap := map[string]string{
		"P1": "adb_binary",
		"P2": "adb_server",
		"P3": "data_dir",
		"P4": "rules_pack",
		"P5": "work_dir",
		"P6": "keychain",
	}
	mappedID := c.ID
	if val, ok := idMap[c.ID]; ok {
		mappedID = val
	}

	var sourceURL *string
	if (c.ID == "P1" || mappedID == "adb_binary") && (c.Status == preflight.StatusMissing || c.Status == preflight.StatusFailed) {
		u := preflight.PlatformToolsDownloadURL()
		sourceURL = &u
	}

	return contract.PreflightCheck{
		ID:        mappedID,
		Title:     c.Title,
		Status:    string(c.Status),
		Detail:    c.Description,
		Remedy:    c.Remedy,
		SourceURL: sourceURL,
	}
}

// --- 2. Device Service ---

func (a *App) DeviceGetConnected() (*contract.Device, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	active := a.tracker.ActiveDevice()
	if active == nil || active.State != adb.StateReady {
		pollCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		_, _ = a.tracker.PollOnce(pollCtx)
		cancel()
		active = a.tracker.ActiveDevice()
	}

	if active == nil {
		return nil, nil
	}
	return a.convertDevice(*active), nil
}

func (a *App) DeviceList() ([]contract.Device, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	pollCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	_, _ = a.tracker.PollOnce(pollCtx)
	cancel()

	devs := a.tracker.Devices()
	res := make([]contract.Device, 0, len(devs))
	for _, d := range devs {
		devDto := a.convertDevice(d)
		if devDto != nil {
			res = append(res, *devDto)
		}
	}
	return res, nil
}

func (a *App) DeviceSelect(serial string) error {
	a.mu.Lock()
	a.activeSerial = serial
	a.mu.Unlock()
	a.tracker.SetActiveSerial(serial)
	return nil
}

// DeviceDiagnoseConnection performs comprehensive connection and USB hardware diagnostics.
func (a *App) DeviceDiagnoseConnection() (*adb.ConnectionDiagnosticResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	a.mu.RLock()
	detector := a.usbDetector
	a.mu.RUnlock()
	if detector == nil {
		detector = adb.DefaultUSBDetector()
	}

	diag, err := a.adbMgr.DiagnoseConnection(ctx, detector)
	if err != nil {
		return nil, wrapADBError(err)
	}
	return diag, nil
}

// AppDiagnoseConnection is an alias for DeviceDiagnoseConnection.
func (a *App) AppDiagnoseConnection() (*adb.ConnectionDiagnosticResult, error) {
	return a.DeviceDiagnoseConnection()
}

// DeviceWirelessPair executes `adb pair <host>:<pairingPort> <pairingCode>` for Android 11+ (API 30+).
func (a *App) DeviceWirelessPair(host string, pairingPort int, pairingCode string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := a.adbMgr.PairWireless(ctx, host, pairingPort, pairingCode); err != nil {
		return wrapADBError(err)
	}
	return nil
}

// AppWirelessPair is an alias for DeviceWirelessPair.
func (a *App) AppWirelessPair(host string, pairingPort int, pairingCode string) error {
	return a.DeviceWirelessPair(host, pairingPort, pairingCode)
}

// DeviceWirelessConnect executes `adb connect <host>:<connectionPort>` to establish an active ADB session.
func (a *App) DeviceWirelessConnect(host string, connectionPort int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	endpoint, err := a.adbMgr.ConnectWireless(ctx, host, connectionPort)
	if err != nil {
		return "", wrapADBError(err)
	}

	// Trigger immediate device polling
	go func() {
		pCtx, pCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer pCancel()
		_, _ = a.tracker.PollOnce(pCtx)
	}()

	return endpoint, nil
}

// AppWirelessConnect is an alias for DeviceWirelessConnect.
func (a *App) AppWirelessConnect(host string, connectionPort int) (string, error) {
	return a.DeviceWirelessConnect(host, connectionPort)
}

// DeviceWirelessDisconnect executes `adb disconnect <host>:<connectionPort>`.
func (a *App) DeviceWirelessDisconnect(host string, connectionPort int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := a.adbMgr.DisconnectWireless(ctx, host, connectionPort); err != nil {
		return wrapADBError(err)
	}

	go func() {
		pCtx, pCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer pCancel()
		_, _ = a.tracker.PollOnce(pCtx)
	}()

	return nil
}

// DeviceSoftReconnect attempts to reset the ADB connection socket without restarting the daemon.
func (a *App) DeviceSoftReconnect(serial string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := a.adbMgr.SoftReconnect(ctx, serial); err != nil {
		return wrapADBError(err)
	}

	go func() {
		pCtx, pCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer pCancel()
		_, _ = a.tracker.PollOnce(pCtx)
	}()

	return nil
}

// AppADBSoftReconnect is an alias for DeviceSoftReconnect.
func (a *App) AppADBSoftReconnect(serial string) error {
	return a.DeviceSoftReconnect(serial)
}

// DeviceRestartServer cleanly kills and restarts the local ADB server daemon.
// Requires explicit user confirmation in the UI before execution.
func (a *App) DeviceRestartServer() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := a.adbMgr.KillServer(ctx); err != nil {
		a.logger.Warn("kill-server returned error during restart", "err", err)
	}

	if err := a.adbMgr.StartServer(ctx); err != nil {
		return wrapADBError(err)
	}

	go func() {
		pCtx, pCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer pCancel()
		_, _ = a.tracker.PollOnce(pCtx)
	}()

	return nil
}

// AppADBRestartServer is an alias for DeviceRestartServer.
func (a *App) AppADBRestartServer() error {
	return a.DeviceRestartServer()
}

func (a *App) convertDevice(d adb.DeviceInfo) *contract.Device {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	manufacturer := "Unknown"
	androidVer := "Unknown"
	sdkInt := 0
	batteryLevel := 0
	isCharging := false

	model := d.Model
	if model == "" {
		model = d.Product
	}

	if d.State == adb.StateReady {
		getpropCtx, cancel1 := context.WithTimeout(ctx, 3*time.Second)
		stdout, _, err := a.runner.Run(getpropCtx, d.Serial, "shell", "getprop")
		cancel1()
		if err == nil {
			props := adb.ParseGetprop(string(stdout))
			if m, ok := props["ro.product.model"]; ok && strings.TrimSpace(m) != "" {
				model = strings.TrimSpace(m)
			} else if m, ok := props["ro.product.marketname"]; ok && strings.TrimSpace(m) != "" {
				model = strings.TrimSpace(m)
			}
			if m, ok := props["ro.product.manufacturer"]; ok && strings.TrimSpace(m) != "" {
				manufacturer = strings.TrimSpace(m)
			} else if b, ok := props["ro.product.brand"]; ok && strings.TrimSpace(b) != "" {
				manufacturer = strings.TrimSpace(b)
			}
			if v, ok := props["ro.build.version.release"]; ok && strings.TrimSpace(v) != "" {
				androidVer = strings.TrimSpace(v)
			}
			if s, ok := props["ro.build.version.sdk"]; ok {
				if num, parseErr := strconv.Atoi(strings.TrimSpace(s)); parseErr == nil {
					sdkInt = num
				}
			}
		}

		batCtx, cancel2 := context.WithTimeout(ctx, 3*time.Second)
		batOut, _, err := a.runner.Run(batCtx, d.Serial, "shell", "dumpsys", "battery")
		cancel2()
		if err == nil {
			bat := adb.ParseBattery(string(batOut))
			batteryLevel = bat.Percent
			isCharging = bat.IsCharging
		}
	} else if d.State == adb.StateUnauthorized {
		manufacturer = "Pending Authorization"
		androidVer = "Pending Authorization"
		if model == "" {
			model = "Unauthorized Android Device"
		}
	} else if d.State == adb.StateOffline {
		manufacturer = "Offline"
		androidVer = "Offline"
		if model == "" {
			model = "Offline Android Device"
		}
	}

	connState := string(d.State)
	if connState != "ready" && connState != "unauthorized" && connState != "offline" {
		connState = "offline"
	}

	if model == "" {
		model = "Unknown Device"
	}

	return &contract.Device{
		ID:              fmt.Sprintf("dev-%s", d.Serial),
		Serial:          d.Serial,
		Model:           model,
		Manufacturer:    manufacturer,
		AndroidVersion:  androidVer,
		SDKInt:          sdkInt,
		BatteryLevel:    batteryLevel,
		IsCharging:      isCharging,
		ConnectionState: connState,
		IsEmulator:      strings.HasPrefix(d.Serial, "emulator-"),
	}
}

// --- 3. Inventory Service ---

func (a *App) getTargetSerial() (string, error) {
	a.mu.RLock()
	serial := a.activeSerial
	a.mu.RUnlock()

	active := a.tracker.ActiveDevice()
	if active == nil {
		return "", wireError("DeviceOffline", "No Android device is currently connected", "Connect device via USB and ensure USB debugging is enabled")
	}

	if active.State == adb.StateUnauthorized {
		return "", wireError("DeviceUnauthorized", "RSA security fingerprint not authorized on device", "Check your phone display and tap 'Allow' or 'Always allow from this computer'")
	}

	if active.State != adb.StateReady {
		return "", wireError("DeviceOffline", fmt.Sprintf("Device is in '%s' state", string(active.State)), "Reconnect device or restart ADB server")
	}

	if serial != "" {
		return serial, nil
	}

	return active.Serial, nil
}

func (a *App) InventoryList(filter contract.InventoryFilter) ([]contract.AppInfo, error) {
	serial, err := a.getTargetSerial()
	if err != nil {
		return nil, err
	}

	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	userOnly := filter.Category != "all" && filter.Category != "system"
	apps, err := a.inventory.ListPackages(ctx, serial, userOnly)
	if err != nil {
		return nil, wireError("AdbMissing", err.Error(), "Ensure ADB is running")
	}

	// Fetch active system states once for batch accuracy
	var overlayMap map[string]bool
	if out, _, runErr := a.runner.Run(ctx, serial, "shell", "dumpsys", "appops"); runErr == nil {
		raw := string(out)
		if strings.TrimSpace(raw) != "" && !strings.Contains(strings.ToLower(raw), "could not find") {
			if parsed, ok := inventory.ParseAppOpsOverlay(raw); ok {
				overlayMap = parsed
			}
		}
	}

	accessMap := make(map[string]bool)
	if out, _, runErr := a.runner.Run(ctx, serial, "shell", "settings", "get", "secure", "enabled_accessibility_services"); runErr == nil {
		raw := strings.TrimSpace(string(out))
		if raw != "" && raw != "null" && raw != "value is null" {
			for _, svc := range strings.Split(raw, ":") {
				parts := strings.Split(svc, "/")
				if len(parts) > 0 && parts[0] != "" {
					accessMap[parts[0]] = true
				}
			}
		}
	}

	var adminMap map[string]bool
	policyOut, _, err1 := a.runner.Run(ctx, serial, "shell", "dumpsys", "device_policy")
	if err1 == nil && strings.TrimSpace(string(policyOut)) != "" && !strings.Contains(strings.ToLower(string(policyOut)), "could not find") {
		parsed := inventory.ParseDeviceAdmins(string(policyOut))
		if len(parsed) > 0 {
			adminMap = parsed
		} else {
			// dumpsys device_policy produced 0 admins; fall back to dpm list-owners
			if dpmOut, _, err2 := a.runner.Run(ctx, serial, "shell", "dpm", "list-owners"); err2 == nil && strings.TrimSpace(string(dpmOut)) != "" && !strings.Contains(strings.ToLower(string(dpmOut)), "could not find") {
				adminMap = inventory.ParseDeviceAdmins(string(dpmOut))
			} else {
				adminMap = parsed // 0 admins confirmed from verified device_policy
			}
		}
	} else {
		// dumpsys failed; fall back to dpm list-owners
		if dpmOut, _, err2 := a.runner.Run(ctx, serial, "shell", "dpm", "list-owners"); err2 == nil && strings.TrimSpace(string(dpmOut)) != "" && !strings.Contains(strings.ToLower(string(dpmOut)), "could not find") {
			adminMap = inventory.ParseDeviceAdmins(string(dpmOut))
		}
		// If neither command succeeded/parsed, adminMap remains nil
	}

	var whitelistMap map[string]bool
	if out, _, runErr := a.runner.Run(ctx, serial, "shell", "dumpsys", "deviceidle", "whitelist"); runErr == nil {
		raw := string(out)
		if strings.TrimSpace(raw) != "" && !strings.Contains(strings.ToLower(raw), "could not find") {
			tempMap := make(map[string]bool)
			for _, line := range strings.Split(raw, "\n") {
				line = strings.TrimSpace(line)
				parts := strings.Split(line, ",")
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p != "" && adb.ValidatePackageName(p) == nil {
						tempMap[p] = true
					}
				}
			}
			if len(tempMap) > 0 {
				whitelistMap = tempMap
			}
		}
	}

	res := make([]contract.AppInfo, 0, len(apps))
	for _, app := range apps {
		if filter.Search != "" {
			if !strings.Contains(strings.ToLower(app.Package), strings.ToLower(filter.Search)) &&
				!strings.Contains(strings.ToLower(app.Label), strings.ToLower(filter.Search)) {
				continue
			}
		}
		res = append(res, a.convertAppInfo(app, overlayMap, accessMap, adminMap, whitelistMap))
	}

	return res, nil
}

func (a *App) InventoryGet(packageName string) (*contract.AppInfo, error) {
	serial, err := a.getTargetSerial()
	if err != nil {
		return nil, err
	}

	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	detail, err := a.inventory.GetPackageDetail(ctx, serial, packageName)
	if err != nil {
		return nil, wireError("ParseFailure", err.Error(), "Check package existence")
	}

	var overlayMap map[string]bool
	if out, _, runErr := a.runner.Run(ctx, serial, "shell", "cmd", "appops", "get", packageName, "SYSTEM_ALERT_WINDOW"); runErr == nil {
		if allowed, ok := inventory.ParseSingleAppOpOverlay(string(out)); ok {
			overlayMap = map[string]bool{packageName: allowed}
		}
	}
	if overlayMap == nil {
		if out2, _, runErr2 := a.runner.Run(ctx, serial, "shell", "appops", "get", packageName, "SYSTEM_ALERT_WINDOW"); runErr2 == nil {
			if allowed, ok := inventory.ParseSingleAppOpOverlay(string(out2)); ok {
				overlayMap = map[string]bool{packageName: allowed}
			}
		}
	}

	accessMap := make(map[string]bool)
	if out, _, runErr := a.runner.Run(ctx, serial, "shell", "settings", "get", "secure", "enabled_accessibility_services"); runErr == nil {
		if strings.Contains(string(out), packageName+"/") {
			accessMap[packageName] = true
		}
	}

	var adminMap map[string]bool
	policyOut, _, err1 := a.runner.Run(ctx, serial, "shell", "dumpsys", "device_policy")
	if err1 == nil && strings.TrimSpace(string(policyOut)) != "" && !strings.Contains(strings.ToLower(string(policyOut)), "could not find") {
		parsed := inventory.ParseDeviceAdmins(string(policyOut))
		if len(parsed) > 0 {
			adminMap = parsed
		} else {
			if dpmOut, _, err2 := a.runner.Run(ctx, serial, "shell", "dpm", "list-owners"); err2 == nil && strings.TrimSpace(string(dpmOut)) != "" && !strings.Contains(strings.ToLower(string(dpmOut)), "could not find") {
				adminMap = inventory.ParseDeviceAdmins(string(dpmOut))
			} else {
				adminMap = parsed
			}
		}
	} else {
		if dpmOut, _, err2 := a.runner.Run(ctx, serial, "shell", "dpm", "list-owners"); err2 == nil && strings.TrimSpace(string(dpmOut)) != "" && !strings.Contains(strings.ToLower(string(dpmOut)), "could not find") {
			adminMap = inventory.ParseDeviceAdmins(string(dpmOut))
		}
	}

	var whitelistMap map[string]bool
	if out, _, runErr := a.runner.Run(ctx, serial, "shell", "dumpsys", "deviceidle", "is-whitelisted", packageName); runErr == nil {
		trimmed := strings.TrimSpace(string(out))
		if trimmed == "true" {
			whitelistMap = map[string]bool{packageName: true}
		} else if trimmed == "false" {
			whitelistMap = map[string]bool{packageName: false}
		}
	}

	appDto := a.convertAppInfo(*detail, overlayMap, accessMap, adminMap, whitelistMap)
	return &appDto, nil
}

func (a *App) InventoryExtractApk(packageName string, targetPath string) (string, error) {
	serial, err := a.getTargetSerial()
	if err != nil {
		return "", err
	}

	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	destDir := targetPath
	if destDir == "" {
		destDir = filepath.Join(a.workDir, "extracted")
	}

	basePath, _, _, err := a.puller.Pull(ctx, serial, packageName, destDir)
	if err != nil {
		return "", wireError("PermissionDenied", err.Error(), "Cannot pull APK")
	}

	return basePath, nil
}

func (a *App) convertAppInfo(app inventory.AppInfo, activeOverlayMap map[string]bool, activeAccessMap map[string]bool, activeAdminMap map[string]bool, batteryWhitelistMap map[string]bool) contract.AppInfo {
	installedAt := ""
	if !app.FirstInstall.IsZero() {
		installedAt = app.FirstInstall.UTC().Format(time.RFC3339)
	}
	updatedAt := ""
	if !app.LastUpdate.IsZero() {
		updatedAt = app.LastUpdate.UTC().Format(time.RFC3339)
	}

	var perms []string
	if len(app.RequestedPermissions) > 0 {
		perms = app.RequestedPermissions
	} else {
		perms = []string{}
	}

	var dangPerms []string
	for _, p := range perms {
		if strings.Contains(p, "SYSTEM_ALERT_WINDOW") ||
			strings.Contains(p, "BIND_ACCESSIBILITY_SERVICE") ||
			strings.Contains(p, "BIND_DEVICE_ADMIN") ||
			strings.Contains(p, "RECEIVE_BOOT_COMPLETED") {
			dangPerms = append(dangPerms, p)
		}
	}
	if len(dangPerms) == 0 {
		dangPerms = []string{}
	}

	// Active overlay check: verified via AppOps
	hasOverlay := false
	if activeOverlayMap != nil {
		hasOverlay = activeOverlayMap[app.Package]
	}

	hasAccess := false
	if activeAccessMap != nil {
		hasAccess = activeAccessMap[app.Package]
	}

	isDevAdmin := false
	if activeAdminMap != nil {
		isDevAdmin = activeAdminMap[app.Package]
	}

	isBatteryOptimized := false
	if batteryWhitelistMap != nil {
		isBatteryOptimized = !batteryWhitelistMap[app.Package]
	}

	sigs := app.CertSHA256
	if len(sigs) == 0 {
		sigs = []string{}
	}

	flags := []string{}
	if !app.IsSystem && (app.Installer == nil || *app.Installer == "com.android.packageinstaller") {
		flags = append(flags, "sideloaded")
	}
	if hasOverlay {
		flags = append(flags, "overlay_granted")
	} else if activeOverlayMap == nil {
		hasOverlayRequested := false
		for _, p := range perms {
			if strings.Contains(p, "SYSTEM_ALERT_WINDOW") {
				hasOverlayRequested = true
				break
			}
		}
		if hasOverlayRequested {
			flags = append(flags, "overlay_status_unknown")
		}
	}
	if hasAccess {
		flags = append(flags, "accessibility_active")
	}
	if isDevAdmin {
		flags = append(flags, "device_admin_active")
	} else if activeAdminMap == nil {
		flags = append(flags, "device_admin_unknown")
	}
	if batteryWhitelistMap != nil {
		if batteryWhitelistMap[app.Package] {
			flags = append(flags, "battery_unrestricted")
		}
	} else {
		flags = append(flags, "battery_status_unknown")
	}

	return contract.AppInfo{
		Package:                    app.Package,
		Label:                      app.Label,
		VersionName:                app.VersionName,
		VersionCode:                fmt.Sprintf("%d", app.VersionCode),
		IsSystem:                   app.IsSystem,
		IsSideloaded:               !app.IsSystem && (app.Installer == nil || *app.Installer != "com.android.vending"),
		InstallerPackage:           app.Installer,
		TargetSDK:                  app.TargetSDK,
		MinSDK:                     app.MinSDK,
		InstalledAt:                installedAt,
		UpdatedAt:                  updatedAt,
		Permissions:                perms,
		DangerousPermissions:       dangPerms,
		HasOverlayPermission:       hasOverlay,
		HasAccessibilityPermission: hasAccess,
		IsDeviceAdmin:              isDevAdmin,
		IsBatteryOptimized:         isBatteryOptimized,
		Signatures:                 sigs,
		SuspiciousFlags:            flags,
	}
}

// --- 4. Scan Service ---

func (a *App) ScanStart(options contract.ScanOptions) (string, error) {
	serial, err := a.getTargetSerial()
	if err != nil {
		return "", err
	}

	if a.scanner == nil {
		return "", wireError("InternalError", "Threat scanner engine not initialized", "Restart application")
	}

	scanCtx, cancel := context.WithCancel(context.Background())
	startTime := time.Now().UTC()
	scanID := fmt.Sprintf("scan-%d", startTime.UnixNano())

	a.mu.Lock()
	a.activeScans[scanID] = cancel
	a.mu.Unlock()

	go func() {
		defer func() {
			a.mu.Lock()
			delete(a.activeScans, scanID)
			a.mu.Unlock()
		}()

		engineOpts := engine.ScanOptions{
			ScanID:        scanID,
			Mode:          options.Mode,
			IncludeSystem: options.IncludeSystem,
		}

		res, scanErr := a.scanner.Scan(scanCtx, serial, engineOpts, func(prog engine.ScanProgress) {
			if a.ctx != nil {
				progDto := contract.ScanProgress{
					Step:           prog.Step,
					Progress:       prog.Progress,
					CurrentPackage: prog.CurrentPackage,
				}
				wailsRuntime.EventsEmit(a.ctx, "scan:progress", progDto)
			}
		})

		if scanErr != nil {
			a.logger.Error("scan failed", "scanId", scanID, "err", scanErr)
			now := time.Now().UTC().Format(time.RFC3339)
			status := "failed"
			if errors.Is(scanErr, context.Canceled) || scanCtx.Err() != nil {
				status = "cancelled"
			}
			failedResult := contract.ScanResult{
				ID:          scanID,
				DeviceID:    fmt.Sprintf("dev-%s", serial),
				Mode:        options.Mode,
				StartedAt:   startTime.Format(time.RFC3339),
				CompletedAt: &now,
				Status:      status,
				Findings:    []contract.Finding{},
				Summary: contract.SecuritySummary{
					Counts: map[string]int{
						"critical": 0, "high": 0, "medium": 0, "low": 0, "info": 0,
					},
					TotalApps:  0,
					LastScanAt: &now,
					Verdict:    "review",
				},
			}
			a.mu.Lock()
			a.cachedResults[scanID] = &failedResult
			a.mu.Unlock()

			if a.ctx != nil {
				wailsRuntime.EventsEmit(a.ctx, "scan:failed", scanErr.Error())
				wailsRuntime.EventsEmit(a.ctx, "scan:done", failedResult)
			}
			return
		}

		contractResult := a.convertEngineScanResult(res)
		a.mu.Lock()
		a.cachedResults[scanID] = &contractResult
		a.mu.Unlock()

		if a.ctx != nil {
			for _, f := range contractResult.Findings {
				wailsRuntime.EventsEmit(a.ctx, "scan:finding", f)
			}
			wailsRuntime.EventsEmit(a.ctx, "scan:done", contractResult)
		}
	}()

	return scanID, nil
}

func (a *App) ScanStatus(scanId string) (*contract.ScanResult, error) {
	a.mu.RLock()
	cached, exists := a.cachedResults[scanId]
	a.mu.RUnlock()

	if exists {
		return cached, nil
	}

	return a.ScanGet(scanId)
}

func (a *App) ScanCancel(scanId string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if cancel, exists := a.activeScans[scanId]; exists {
		cancel()
		delete(a.activeScans, scanId)
	}
	return nil
}

func (a *App) ScanGet(scanId string) (*contract.ScanResult, error) {
	if a.store == nil {
		return nil, wireError("InternalError", "Database store unavailable", "Check database access")
	}

	record, err := a.store.GetScanResult(scanId)
	if err != nil {
		return nil, wireError("ParseFailure", "Scan not found", "Verify scan ID")
	}

	dto := a.convertStoreScanRecord(record)
	return &dto, nil
}

func (a *App) ScanQuick() (*contract.ScanResult, error) {
	scanId, err := a.ScanStart(contract.ScanOptions{Mode: "quick", IncludeSystem: false})
	if err != nil {
		return nil, err
	}

	// Poll until terminal state (completed, partial, failed, cancelled) for synchronous callers
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(300 * time.Millisecond)
		res, _ := a.ScanStatus(scanId)
		if res != nil {
			switch res.Status {
			case "completed", "partial", "failed", "cancelled":
				return res, nil
			}
		}
	}

	return nil, wireError("Timeout", "Quick scan execution timed out", "Retry scan")
}

func mapEvidenceType(source string) string {
	evType := "manifest"
	srcLower := strings.ToLower(source)
	if strings.Contains(srcLower, "permission") {
		evType = "permission"
	} else if strings.Contains(srcLower, "intel") || strings.Contains(srcLower, "hash") {
		evType = "hash"
	} else if strings.Contains(srcLower, "class") {
		evType = "class_prefix"
	} else if strings.Contains(srcLower, "log") {
		evType = "log"
	} else if strings.Contains(srcLower, "setting") {
		evType = "setting"
	}
	return evType
}

func (a *App) convertEngineScanResult(res *engine.ScanResult) contract.ScanResult {
	var findings []contract.Finding
	counts := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
		"info":     0,
	}

	for _, f := range res.Findings {
		sev := strings.ToLower(f.Severity)
		counts[sev]++

		var evs []contract.Evidence
		for _, e := range f.Evidence {
			evs = append(evs, contract.Evidence{
				Type:    mapEvidenceType(e.Source),
				Title:   e.Source,
				Details: e.Snippet,
			})
		}

		pkgStr := f.Package
		var pkgPtr *string
		if pkgStr != "" {
			pkgPtr = &pkgStr
		}

		var recAct *contract.RemediationAction
		if len(f.Actions) > 0 {
			act := a.remediate.CreateAction(f.Actions[0], f.Package, "", "", false)
			recAct = &contract.RemediationAction{
				ID:            act.ID,
				Type:          act.Type,
				Package:       act.Package,
				Title:         act.Title,
				Impact:        act.Impact,
				Undoable:      act.Undoable,
				IsBlocked:     act.IsBlocked,
				BlockedReason: act.BlockedReason,
			}
		}

		findings = append(findings, contract.Finding{
			ID:                f.ID,
			RuleID:            f.RuleID,
			Title:             f.Title,
			Severity:          sev,
			Confidence:        f.Confidence,
			Package:           pkgPtr,
			WhyFlagged:        f.WhyFlagged,
			FalsePositiveHint: f.FalsePositiveHint,
			Evidence:          evs,
			RecommendedAction: recAct,
		})
	}

	verdict := "clean"
	if counts["critical"] > 0 || counts["high"] > 0 {
		verdict = "high_risk"
	} else if counts["medium"] > 0 {
		verdict = "review"
	} else if res.Partial {
		verdict = "review"
	}

	compAt := res.FinishedAt.UTC().Format(time.RFC3339)

	totalApps := res.TotalApps
	if totalApps == 0 {
		pkgs := make(map[string]struct{})
		for _, f := range res.Findings {
			if f.Package != "" {
				pkgs[f.Package] = struct{}{}
			}
		}
		totalApps = len(pkgs)
	}

	status := "completed"
	if res.Partial {
		status = "partial"
	}

	return contract.ScanResult{
		ID:          res.ID,
		DeviceID:    fmt.Sprintf("dev-%s", res.DeviceSerial),
		Mode:        res.Mode,
		StartedAt:   res.StartedAt.UTC().Format(time.RFC3339),
		CompletedAt: &compAt,
		Status:      status,
		Findings:    findings,
		Summary: contract.SecuritySummary{
			Counts:     counts,
			TotalApps:  totalApps,
			LastScanAt: &compAt,
			Verdict:    verdict,
		},
	}
}

func (a *App) convertStoreFinding(f store.FindingRecord) contract.Finding {
	sev := strings.ToLower(f.Severity)
	var evs []contract.Evidence
	for _, e := range f.Evidence {
		evs = append(evs, contract.Evidence{
			Type:    mapEvidenceType(e.Source),
			Title:   e.Source,
			Details: e.Snippet,
		})
	}
	if len(evs) == 0 {
		evs = []contract.Evidence{}
	}

	pkgStr := f.Package
	var pkgPtr *string
	if pkgStr != "" {
		pkgPtr = &pkgStr
	}

	var recAct *contract.RemediationAction
	if len(f.Actions) > 0 && pkgStr != "" {
		actType := f.Actions[0]
		recAct = &contract.RemediationAction{
			ID:       fmt.Sprintf("act-rec-%s", f.ID),
			Type:     actType,
			Package:  pkgStr,
			Title:    fmt.Sprintf("Remediate %s on %s", actType, pkgStr),
			Impact:   "Action recommended to mitigate flagged threat.",
			Undoable: true,
		}
	}

	return contract.Finding{
		ID:                f.ID,
		RuleID:            f.RuleID,
		Title:             f.Title,
		Severity:          sev,
		Confidence:        f.Confidence,
		Package:           pkgPtr,
		WhyFlagged:        f.WhyFlagged,
		FalsePositiveHint: f.FalsePositiveHint,
		Evidence:          evs,
		RecommendedAction: recAct,
	}
}

func (a *App) convertStoreScanRecord(record *store.ScanRecord) contract.ScanResult {
	var findings []contract.Finding
	counts := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
		"info":     0,
	}

	for _, f := range record.Findings {
		sev := strings.ToLower(f.Severity)
		counts[sev]++
		findings = append(findings, a.convertStoreFinding(f))
	}
	if len(findings) == 0 {
		findings = []contract.Finding{}
	}

	verdict := "clean"
	if counts["critical"] > 0 || counts["high"] > 0 {
		verdict = "high_risk"
	} else if counts["medium"] > 0 {
		verdict = "review"
	} else if record.Partial {
		verdict = "review"
	}

	var compAt *string
	if !record.FinishedAt.IsZero() {
		s := record.FinishedAt.UTC().Format(time.RFC3339)
		compAt = &s
	}

	totalApps := record.TotalApps
	if totalApps == 0 {
		pkgs := make(map[string]struct{})
		for _, f := range record.Findings {
			if f.Package != "" {
				pkgs[f.Package] = struct{}{}
			}
		}
		totalApps = len(pkgs)
	}

	status := "completed"
	if record.Partial {
		status = "partial"
	}

	return contract.ScanResult{
		ID:          record.ID,
		DeviceID:    fmt.Sprintf("dev-%s", record.DeviceSerial),
		Mode:        record.Mode,
		StartedAt:   record.StartedAt.UTC().Format(time.RFC3339),
		CompletedAt: compAt,
		Status:      status,
		Findings:    findings,
		Summary: contract.SecuritySummary{
			Counts:     counts,
			TotalApps:  totalApps,
			LastScanAt: compAt,
			Verdict:    verdict,
		},
	}
}

// --- 5. Monitor Service ---

func (a *App) MonitorStart() error {
	serial, err := a.getTargetSerial()
	if err != nil {
		return err
	}

	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	return a.monitor.Start(ctx, serial)
}

func (a *App) MonitorStop() error {
	a.monitor.Stop()
	return nil
}

func (a *App) MonitorMarkPopup() error {
	a.monitor.MarkUserPopup(time.Now().UTC())
	return nil
}

func (a *App) MonitorSnapshot() (contract.MonitorSnapshot, error) {
	snap := a.monitor.Snapshot()

	var events []contract.MonitorEvent
	for _, e := range snap.Events {
		events = append(events, contract.MonitorEvent{
			ID:        e.ID,
			Timestamp: e.Timestamp,
			Type:      string(e.Type),
			Package:   e.Package,
			Details:   e.Details,
		})
	}

	var candidates []contract.MonitorCandidate
	for _, c := range snap.Candidates {
		var candEvents []contract.MonitorEvent
		for _, ce := range c.TimelineEvents {
			candEvents = append(candEvents, contract.MonitorEvent{
				ID:        ce.ID,
				Timestamp: ce.Timestamp,
				Type:      string(ce.Type),
				Package:   ce.Package,
				Details:   ce.Details,
			})
		}
		candidates = append(candidates, contract.MonitorCandidate{
			Package:        c.Package,
			AppLabel:       c.AppLabel,
			Confidence:     c.Confidence,
			Score:          c.Score,
			TimelineEvents: candEvents,
			WhyFlagged:     c.WhyFlagged,
		})
	}

	return contract.MonitorSnapshot{
		Status: contract.MonitorStatus{
			IsActive:      snap.Status.IsActive,
			StartedAt:     snap.Status.StartedAt,
			SourcesActive: snap.Status.SourcesActive,
		},
		Events:     events,
		Candidates: candidates,
	}, nil
}

// --- 6. Audit Service ---

func (a *App) AuditRun() ([]contract.AuditItem, error) {
	serial, err := a.getTargetSerial()
	if err != nil {
		return nil, err
	}

	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	items, err := a.auditSvc.RunAudit(ctx, serial)
	if err != nil {
		return nil, wireError("InternalError", err.Error(), "Ensure device is responsive")
	}

	dtos := make([]contract.AuditItem, 0, len(items))
	for _, it := range items {
		statusStr := string(it.Status)
		if statusStr == "critical" {
			statusStr = "issue"
		} else if statusStr == "not_available" {
			statusStr = "unavailable"
		}

		dtos = append(dtos, contract.AuditItem{
			ID:       it.ID,
			Category: string(it.Category),
			Title:    it.Title,
			Status:   statusStr,
			Summary:  it.Description,
			Evidence: []string{
				fmt.Sprintf("Current value: %s", it.CurrentValue),
				fmt.Sprintf("Recommended value: %s", it.RecommendedValue),
			},
			Remedy: it.Remediation,
		})
	}

	a.mu.Lock()
	a.latestAudit = dtos
	a.mu.Unlock()

	return dtos, nil
}

func (a *App) AuditGetLatest() ([]contract.AuditItem, error) {
	a.mu.RLock()
	latest := a.latestAudit
	a.mu.RUnlock()

	if len(latest) > 0 {
		return latest, nil
	}
	return a.AuditRun()
}

// --- 7. Remediate Service ---

func (a *App) RemediateCreatePlan(findingIds []string) (contract.RemediationPlan, error) {
	serial, err := a.getTargetSerial()
	if err != nil {
		return contract.RemediationPlan{}, err
	}
	if serial == "" {
		return contract.RemediationPlan{}, wireError("DeviceOffline", "No active trusted device connected", "Connect a device via USB and verify ADB authorization")
	}

	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	var actions []contract.RemediationAction
	requiresBackup := false

	a.mu.Lock()
	defer a.mu.Unlock()

	expectedDevID := fmt.Sprintf("dev-%s", serial)

	for _, fid := range findingIds {
		var pkg string
		actionType := remediate.ActionDisable
		var perm string
		var appop string

		// 1. Look in cached scan results
		var foundFinding *contract.Finding
		for _, res := range a.cachedResults {
			if res == nil || res.DeviceID != expectedDevID {
				continue
			}
			for _, f := range res.Findings {
				if f.ID == fid {
					fCopy := f
					foundFinding = &fCopy
					break
				}
			}
			if foundFinding != nil {
				break
			}
		}

		if foundFinding != nil {
			if foundFinding.Package != nil {
				pkg = *foundFinding.Package
			}
			if foundFinding.RecommendedAction != nil {
				actionType = foundFinding.RecommendedAction.Type
			}
			for _, ev := range foundFinding.Evidence {
				if ev.Type == "permission" && strings.Contains(ev.Details, "permission.") {
					cleaned := extractCleanPermission(ev.Details)
					if cleaned != "" {
						perm = cleaned
					}
				}
			}
		} else if a.store != nil {
			// 2. Look in database store
			record, err := a.store.GetFinding(fid)
			if err == nil && record != nil {
				scanRec, scanErr := a.store.GetScanResult(record.ScanID)
				if scanErr != nil || scanRec == nil || scanRec.DeviceSerial != serial {
					scanDev := ""
					if scanRec != nil {
						scanDev = scanRec.DeviceSerial
					}
					a.logger.Warn("remediate: finding does not belong to active device", "findingId", fid, "scanDevice", scanDev, "activeDevice", serial)
					continue
				}

				pkg = record.Package
				if len(record.Actions) > 0 {
					actionType = record.Actions[0]
				}
				for _, ev := range record.Evidence {
					if strings.Contains(ev.Snippet, "permission.") {
						cleaned := extractCleanPermission(ev.Snippet)
						if cleaned != "" {
							perm = cleaned
						}
					}
				}
			}
		}

		if pkg == "" {
			a.logger.Warn("remediate: finding not found or missing package", "findingId", fid)
			continue
		}

		if actionType == remediate.ActionRevokePermission && perm == "" {
			actionType = remediate.ActionDisable
		}
		if actionType == remediate.ActionClearAppop {
			if strings.Contains(perm, "SYSTEM_ALERT_WINDOW") {
				appop = "SYSTEM_ALERT_WINDOW"
			}
			if appop == "" {
				actionType = remediate.ActionDisable
			}
		}

		// Detect if system package via trusted Package Manager metadata
		isSystem := a.isSystemPackage(ctx, serial, pkg)

		act := a.remediate.CreateAction(actionType, pkg, perm, appop, isSystem)
		act.DeviceSerial = serial
		a.cachedActions[act.ID] = act

		if act.Type == remediate.ActionUninstall && !isSystem {
			requiresBackup = true
		}

		actions = append(actions, contract.RemediationAction{
			ID:            act.ID,
			Type:          act.Type,
			Package:       act.Package,
			Title:         act.Title,
			Impact:        act.Impact,
			Undoable:      act.Undoable,
			IsBlocked:     act.IsBlocked,
			BlockedReason: act.BlockedReason,
		})
	}

	return contract.RemediationPlan{
		Actions:        actions,
		RequiresBackup: requiresBackup,
	}, nil
}

func (a *App) RemediateExecute(plan contract.RemediationPlan) ([]contract.ActionLogEntry, error) {
	serial, err := a.getTargetSerial()
	if err != nil {
		return nil, err
	}
	if serial == "" {
		return nil, wireError("DeviceOffline", "No active trusted device connected", "Connect a device via USB and verify ADB authorization")
	}

	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	// Atomic validation and reservation of all actions under exclusive lock
	a.mu.Lock()

	// 1. Check for duplicate action IDs in plan
	seenIDs := make(map[string]struct{}, len(plan.Actions))
	for _, actDto := range plan.Actions {
		if _, seen := seenIDs[actDto.ID]; seen {
			a.mu.Unlock()
			a.logger.Warn("remediate: duplicate action ID in execution plan", "actionId", actDto.ID, "package", actDto.Package)
			return nil, wireError("InvalidAction", fmt.Sprintf("Duplicate action ID '%s' in plan.", actDto.ID), "Ensure each remediation action is submitted only once.")
		}
		seenIDs[actDto.ID] = struct{}{}
	}

	// 2. Validate all actions against cachedActions and device serial
	reservedActions := make([]remediate.Action, 0, len(plan.Actions))
	for _, actDto := range plan.Actions {
		cachedAct, exists := a.cachedActions[actDto.ID]
		if !exists {
			a.mu.Unlock()
			a.logger.Warn("remediate: unrecognized or expired action ID in plan", "actionId", actDto.ID, "package", actDto.Package)
			return nil, wireError("InvalidAction", fmt.Sprintf("Action ID '%s' is unrecognized or expired. Actions must originate from a verified backend plan.", actDto.ID), "Re-generate the remediation plan before executing.")
		}

		if cachedAct.DeviceSerial == "" || cachedAct.DeviceSerial != serial {
			a.mu.Unlock()
			a.logger.Warn("remediate: plan device mismatch", "actionId", actDto.ID, "planDevice", cachedAct.DeviceSerial, "activeDevice", serial)
			return nil, wireError("DeviceMismatch", fmt.Sprintf("Action ID '%s' was planned for device '%s', but active device is '%s'.", actDto.ID, cachedAct.DeviceSerial, serial), "Re-generate the remediation plan for the active device before executing.")
		}

		reservedActions = append(reservedActions, cachedAct)
	}

	// 3. Atomically consume / delete all action IDs from cachedActions
	for _, act := range reservedActions {
		delete(a.cachedActions, act.ID)
	}

	a.mu.Unlock()

	// Phase 2: Execution using verified canonical backend actions
	results := make([]contract.ActionLogEntry, 0, len(reservedActions))
	for _, cachedAct := range reservedActions {
		canonicalActDto := contract.RemediationAction{
			ID:            cachedAct.ID,
			Type:          cachedAct.Type,
			Package:       cachedAct.Package,
			Title:         cachedAct.Title,
			Impact:        cachedAct.Impact,
			Undoable:      cachedAct.Undoable,
			IsBlocked:     cachedAct.IsBlocked,
			BlockedReason: cachedAct.BlockedReason,
		}

		logEntry, execErr := a.remediate.Execute(ctx, serial, cachedAct)
		if logEntry != nil {
			results = append(results, contract.ActionLogEntry{
				ID:         logEntry.ID,
				Timestamp:  logEntry.Timestamp,
				Action:     canonicalActDto,
				Status:     logEntry.Status,
				Error:      logEntry.Error,
				BackupPath: logEntry.BackupPath,
				AuditSaved: logEntry.AuditSaved,
				AuditError: logEntry.AuditError,
			})
		} else if execErr != nil {
			errStr := execErr.Error()
			auditSaved := false
			auditErr := "action execution failed before audit"
			results = append(results, contract.ActionLogEntry{
				ID:         fmt.Sprintf("log-%d", time.Now().UnixNano()),
				Timestamp:  time.Now().UTC().Format(time.RFC3339),
				Action:     canonicalActDto,
				Status:     "failed",
				Error:      &errStr,
				AuditSaved: &auditSaved,
				AuditError: &auditErr,
			})
		}
		if execErr != nil {
			a.logger.Error("action execution failed", "action", cachedAct.ID, "err", execErr)
		}
	}

	return results, nil
}

func (a *App) RemediateGetHistory() ([]contract.ActionLogEntry, error) {
	serial, _ := a.getTargetSerial()
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	history, err := a.remediate.GetHistory(ctx, serial)
	if err != nil {
		return nil, err
	}

	dtos := make([]contract.ActionLogEntry, 0, len(history))
	for _, h := range history {
		dtos = append(dtos, contract.ActionLogEntry{
			ID:        h.ID,
			Timestamp: h.Timestamp,
			Action: contract.RemediationAction{
				ID:        h.Action.ID,
				Type:      h.Action.Type,
				Package:   h.Action.Package,
				Title:     h.Action.Title,
				Impact:    h.Action.Impact,
				Undoable:  h.Action.Undoable,
				IsBlocked: h.Action.IsBlocked,
			},
			Status:     h.Status,
			Error:      h.Error,
			BackupPath: h.BackupPath,
			AuditSaved: h.AuditSaved,
			AuditError: h.AuditError,
		})
	}

	return dtos, nil
}

func (a *App) RemediateUndo(logEntryId string) (contract.ActionLogEntry, error) {
	serial, err := a.getTargetSerial()
	if err != nil {
		return contract.ActionLogEntry{}, err
	}

	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	entry, err := a.remediate.Undo(ctx, serial, logEntryId)
	if err != nil {
		return contract.ActionLogEntry{}, wireError("InternalError", err.Error(), "Verify undo capability")
	}

	return contract.ActionLogEntry{
		ID:        entry.ID,
		Timestamp: entry.Timestamp,
		Action: contract.RemediationAction{
			ID:       entry.Action.ID,
			Type:     entry.Action.Type,
			Package:  entry.Action.Package,
			Title:    entry.Action.Title,
			Undoable: false,
		},
		Status:     "undone",
		Error:      entry.Error,
		BackupPath: entry.BackupPath,
		AuditSaved: entry.AuditSaved,
		AuditError: entry.AuditError,
	}, nil
}

func (a *App) isSystemPackage(ctx context.Context, serial, pkg string) bool {
	if a.inventory != nil {
		return a.inventory.IsSystemPackage(ctx, serial, pkg)
	}
	if a.runner != nil {
		return inventory.NewService(a.runner, a.logger).IsSystemPackage(ctx, serial, pkg)
	}
	return false
}

// --- 8. Report Service ---

func (a *App) ReportList() ([]contract.ScanSummaryItem, error) {
	serial, _ := a.getTargetSerial()
	records, err := a.store.ListScanSummaries(serial)
	if err != nil {
		return nil, err
	}

	res := make([]contract.ScanSummaryItem, 0, len(records))
	for _, r := range records {
		var compAt *string
		if !r.FinishedAt.IsZero() {
			s := r.FinishedAt.UTC().Format(time.RFC3339)
			compAt = &s
		}
		counts := r.Counts
		if counts == nil {
			counts = map[string]int{
				"critical": 0,
				"high":     r.HighSeverityCount,
				"medium":   0,
				"low":      0,
				"info":     0,
			}
		}
		verdict := r.Verdict
		if verdict == "" {
			verdict = "clean"
			if r.HighSeverityCount > 0 {
				verdict = "high_risk"
			}
		}
		res = append(res, contract.ScanSummaryItem{
			ID:            r.ID,
			StartedAt:     r.StartedAt.UTC().Format(time.RFC3339),
			CompletedAt:   compAt,
			Mode:          r.Mode,
			Verdict:       verdict,
			FindingsCount: counts,
		})
	}

	return res, nil
}

func (a *App) ReportGet(id string) (*contract.ScanResult, error) {
	return a.ScanGet(id)
}

func (a *App) ReportCompare(scanIdA string, scanIdB string) (contract.ScanComparison, error) {
	recA, err := a.store.GetScanResult(scanIdA)
	if err != nil {
		return contract.ScanComparison{}, err
	}
	recB, err := a.store.GetScanResult(scanIdB)
	if err != nil {
		return contract.ScanComparison{}, err
	}

	comp := report.CompareScans(*recA, *recB)

	itemA := contract.ScanSummaryItem{
		ID:            comp.ScanA.ID,
		StartedAt:     comp.ScanA.StartedAt,
		CompletedAt:   comp.ScanA.CompletedAt,
		Mode:          comp.ScanA.Mode,
		Verdict:       comp.ScanA.Verdict,
		FindingsCount: comp.ScanA.FindingsCount,
	}
	itemB := contract.ScanSummaryItem{
		ID:            comp.ScanB.ID,
		StartedAt:     comp.ScanB.StartedAt,
		CompletedAt:   comp.ScanB.CompletedAt,
		Mode:          comp.ScanB.Mode,
		Verdict:       comp.ScanB.Verdict,
		FindingsCount: comp.ScanB.FindingsCount,
	}

	var newFindings []contract.Finding
	for _, f := range comp.NewFindings {
		newFindings = append(newFindings, a.convertStoreFinding(f))
	}
	if len(newFindings) == 0 {
		newFindings = []contract.Finding{}
	}

	var resolvedFindings []contract.Finding
	for _, f := range comp.ResolvedFindings {
		resolvedFindings = append(resolvedFindings, a.convertStoreFinding(f))
	}
	if len(resolvedFindings) == 0 {
		resolvedFindings = []contract.Finding{}
	}

	var unchangedFindings []contract.Finding
	for _, f := range comp.UnchangedFindings {
		unchangedFindings = append(unchangedFindings, a.convertStoreFinding(f))
	}
	if len(unchangedFindings) == 0 {
		unchangedFindings = []contract.Finding{}
	}

	return contract.ScanComparison{
		ScanA:             itemA,
		ScanB:             itemB,
		NewFindings:       newFindings,
		ResolvedFindings:  resolvedFindings,
		UnchangedFindings: unchangedFindings,
	}, nil
}

func (a *App) ReportExport(options contract.ReportExportOptions) (string, error) {
	rec, err := a.store.GetScanResult(options.ScanID)
	if err != nil {
		return "", err
	}

	targetPath := ""
	if options.TargetPath != nil && *options.TargetPath != "" {
		targetPath = *options.TargetPath
	} else {
		targetPath = filepath.Join(a.workDir, fmt.Sprintf("report_%s.%s", options.ScanID, options.Format))
	}

	if options.Format == "json" {
		data, err := report.ExportJSON(*rec, false)
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			return "", err
		}
	} else {
		devInfo := report.DeviceInfo{
			Serial:       rec.DeviceSerial,
			Model:        "Unknown Device",
			Manufacturer: "Unknown",
			AndroidVer:   "Unknown",
			SDKInt:       0,
		}
		if a.store != nil {
			if dRec, err := a.store.GetDevice(rec.DeviceSerial); err == nil && dRec != nil {
				if dRec.Model != "" {
					devInfo.Model = dRec.Model
				}
				if dRec.Manufacturer != "" {
					devInfo.Manufacturer = dRec.Manufacturer
				}
				if dRec.AndroidVersion != "" {
					devInfo.AndroidVer = dRec.AndroidVersion
				}
				devInfo.SDKInt = dRec.SDK
			}
		}
		htmlData, err := report.ExportHTML(*rec, devInfo, false)
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(targetPath, []byte(htmlData), 0644); err != nil {
			return "", err
		}
	}

	return targetPath, nil
}

// --- 9. Settings Service ---

func (a *App) SettingsGet() (contract.Settings, error) {
	all, _ := a.store.GetAllSettings()

	theme := "system"
	if t, ok := all["theme"]; ok && t != "" {
		theme = t
	}

	var adbPath *string
	if p, ok := all["adbPath"]; ok && p != "" {
		adbPath = &p
	}

	var customPort *int
	if cp, ok := all["customAdbPort"]; ok && cp != "" {
		if num, err := strconv.Atoi(cp); err == nil {
			customPort = &num
		}
	}

	timeout := 120
	if to, ok := all["deepScanTimeoutSec"]; ok && to != "" {
		if num, err := strconv.Atoi(to); err == nil {
			timeout = num
		}
	}

	autoPref := true
	if ap, ok := all["autoCheckPreflight"]; ok && ap == "false" {
		autoPref = false
	}

	return contract.Settings{
		Theme:              theme,
		AdbPath:            adbPath,
		CustomAdbPort:      customPort,
		DeepScanTimeoutSec: timeout,
		AutoCheckPreflight: autoPref,
	}, nil
}

func (a *App) SettingsUpdate(partial contract.Settings) (contract.Settings, error) {
	if partial.Theme != "" {
		_ = a.store.SetSetting("theme", partial.Theme)
	}
	if partial.AdbPath != nil {
		_ = a.store.SetSetting("adbPath", *partial.AdbPath)
		a.adbMgr.SetCustomPath(*partial.AdbPath)
		if psr, ok := a.runner.(interface{ SetADBPath(string) }); ok {
			psr.SetADBPath(*partial.AdbPath)
		}
	}
	if partial.CustomAdbPort != nil {
		_ = a.store.SetSetting("customAdbPort", fmt.Sprintf("%d", *partial.CustomAdbPort))
	}
	if partial.DeepScanTimeoutSec > 0 {
		_ = a.store.SetSetting("deepScanTimeoutSec", fmt.Sprintf("%d", partial.DeepScanTimeoutSec))
	}
	_ = a.store.SetSetting("autoCheckPreflight", fmt.Sprintf("%t", partial.AutoCheckPreflight))

	return a.SettingsGet()
}

// --- 10. Allowlist Service ---

func (a *App) AllowlistList() ([]contract.AllowlistEntry, error) {
	entries, err := a.store.ListAllowlist()
	if err != nil {
		return nil, err
	}

	dtos := make([]contract.AllowlistEntry, 0, len(entries))
	for _, e := range entries {
		var reason *string
		if e.Reason != "" {
			reason = &e.Reason
		}
		dtos = append(dtos, contract.AllowlistEntry{
			ID:      e.ID,
			Type:    "package",
			Target:  e.Package,
			AddedAt: e.AddedAt.UTC().Format(time.RFC3339),
			Reason:  reason,
		})
	}

	return dtos, nil
}

func (a *App) AllowlistAdd(entry contract.AllowlistEntry) (contract.AllowlistEntry, error) {
	now := time.Now().UTC()
	id := fmt.Sprintf("allow-%d", now.UnixNano())
	reason := ""
	if entry.Reason != nil {
		reason = *entry.Reason
	}

	err := a.store.AddAllowlistEntry(store.AllowlistRecord{
		ID:      id,
		Package: entry.Target,
		Reason:  reason,
		AddedAt: now,
	})
	if err != nil {
		return contract.AllowlistEntry{}, err
	}

	return contract.AllowlistEntry{
		ID:      id,
		Type:    entry.Type,
		Target:  entry.Target,
		AddedAt: now.Format(time.RFC3339),
		Reason:  entry.Reason,
	}, nil
}

func (a *App) AllowlistRemove(id string) error {
	return a.store.RemoveAllowlistEntry(id)
}

// --- 10. ClamAV Service ---

// ClamAVConfig represents the local ClamAV engine configuration.
type ClamAVConfig struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
}

// ClamAVStatus represents the result of testing the ClamAV daemon connection.
type ClamAVStatus struct {
	Connected bool   `json:"connected"`
	Version   string `json:"version,omitempty"`
	Error     string `json:"error,omitempty"`
}

// ClamAVGetConfig retrieves the ClamAV configuration.
func (a *App) ClamAVGetConfig() (ClamAVConfig, error) {
	all, _ := a.store.GetAllSettings()
	enabled := false
	if e, ok := all["clamav_enabled"]; ok && e == "true" {
		enabled = true
	}
	host := intel.DefaultClamAVAddress
	if h, ok := all["clamav_host"]; ok && h != "" {
		if valid, err := intel.ValidateClamAVAddress(h); err == nil {
			host = valid
		}
	}
	return ClamAVConfig{
		Enabled: enabled,
		Host:    host,
	}, nil
}

// ClamAVSaveConfig saves the ClamAV configuration.
func (a *App) ClamAVSaveConfig(enabled bool, host string) error {
	trimmedHost := strings.TrimSpace(host)
	if trimmedHost == "" {
		trimmedHost = intel.DefaultClamAVAddress
	}
	validHost, err := intel.ValidateClamAVAddress(trimmedHost)
	if err != nil {
		return err
	}
	if err := a.store.SetSetting("clamav_enabled", fmt.Sprintf("%t", enabled)); err != nil {
		return err
	}
	if err := a.store.SetSetting("clamav_host", validHost); err != nil {
		return err
	}
	if a.intelAgg != nil {
		if enabled {
			a.intelAgg.SetClamAVClient(intel.NewClamAVClient(validHost))
		} else {
			a.intelAgg.SetClamAVClient(nil)
		}
	}
	return nil
}

// ClamAVTestConnection probes the ClamAV daemon at the specified host.
func (a *App) ClamAVTestConnection(host string) (ClamAVStatus, error) {
	trimmedHost := strings.TrimSpace(host)
	if trimmedHost == "" {
		trimmedHost = intel.DefaultClamAVAddress
	}
	validHost, err := intel.ValidateClamAVAddress(trimmedHost)
	if err != nil {
		return ClamAVStatus{
			Connected: false,
			Error:     err.Error(),
		}, nil
	}
	client := intel.NewClamAVClient(validHost)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ver, err := client.Version(ctx)
	if err != nil {
		return ClamAVStatus{
			Connected: false,
			Error:     err.Error(),
		}, nil
	}
	return ClamAVStatus{
		Connected: true,
		Version:   ver,
	}, nil
}

// AppSystemInfo represents application, rules catalog, and bridge runtime metadata.
type AppSystemInfo struct {
	AppVersion      string `json:"appVersion"`
	RulesVersion    string `json:"rulesVersion"`
	RulesCount      int    `json:"rulesCount"`
	AdbVersion      string `json:"adbVersion,omitempty"`
	ContractVersion string `json:"contractVersion"`
}

// AppGetSystemInfo retrieves dynamic application and rules engine metadata.
func (a *App) AppGetSystemInfo() AppSystemInfo {
	rulesCount := rules.Count()
	adbVer := ""
	if a.adbMgr != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if ver, err := a.adbMgr.CheckVersion(ctx); err == nil {
			adbVer = ver
		}
	}
	return AppSystemInfo{
		AppVersion:      version.AppVersion,
		RulesVersion:    rules.RulesVersion(),
		RulesCount:      rulesCount,
		AdbVersion:      adbVer,
		ContractVersion: version.ContractVersion,
	}
}
