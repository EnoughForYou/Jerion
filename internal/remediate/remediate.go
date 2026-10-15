package remediate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/EnoughForYou/Jerion/internal/adb"
	"github.com/EnoughForYou/Jerion/internal/inventory"
	"github.com/EnoughForYou/Jerion/internal/logx"
	"github.com/EnoughForYou/Jerion/internal/store"
)

// Action types matching contract v1.0.0.
const (
	ActionDisable          = "disable"
	ActionUninstall        = "uninstall"
	ActionRevokePermission = "revoke_permission"
	ActionClearAppop       = "clear_appop"
)

// RemediationAction represents an individual guarded action.
type Action struct {
	ID               string  `json:"id"`
	Type             string  `json:"type"`
	Package          string  `json:"package"`
	Title            string  `json:"title"`
	Impact           string  `json:"impact"`
	Undoable         bool    `json:"undoable"`
	IsBlocked        bool    `json:"isBlocked"`
	BlockedReason    *string `json:"blockedReason,omitempty"`
	TargetPermission string  `json:"targetPermission,omitempty"`
	TargetAppop      string  `json:"targetAppop,omitempty"`
	IsSystemApp      bool    `json:"isSystemApp,omitempty"`
	DeviceSerial     string  `json:"deviceSerial,omitempty"`
}

// RemediationPlan represents an actionable queue of remediation actions.
type Plan struct {
	Actions        []Action `json:"actions"`
	RequiresBackup bool     `json:"requiresBackup"`
}

// ActionLogEntry represents the historical audit log of an executed action.
type ActionLogEntry struct {
	ID         string  `json:"id"`
	Timestamp  string  `json:"timestamp"`
	Action     Action  `json:"action"`
	Status     string  `json:"status"` // "success" | "failed" | "undone"
	Error      *string `json:"error,omitempty"`
	BackupPath *string `json:"backupPath,omitempty"`
}

// Manager coordinates guarded remediation execution, backup, and reversal.
type Manager struct {
	runner    adb.Runner
	store     *store.Store
	logger    *logx.Logger
	backupDir string
}

// NewManager creates a remediation manager.
func NewManager(runner adb.Runner, st *store.Store, logger *logx.Logger, backupDir string) *Manager {
	if logger == nil {
		logger = logx.GetDefault()
	}
	if backupDir == "" {
		backupDir = filepath.Join(os.TempDir(), "jerion_backups")
	}
	return &Manager{
		runner:    runner,
		store:     st,
		logger:    logger,
		backupDir: backupDir,
	}
}

// CreateAction constructs a validated remediation action.
func (m *Manager) CreateAction(actionType, pkg, perm, appop string, isSystem bool) Action {
	isBlocked, reason := CheckGuardrail(pkg)

	title := ""
	impact := ""
	undoable := true

	switch actionType {
	case ActionDisable:
		title = fmt.Sprintf("Disable package %s", pkg)
		impact = "Package execution is suspended and frozen from launching."
		undoable = true
	case ActionUninstall:
		if isSystem {
			title = fmt.Sprintf("Uninstall system package %s for user 0", pkg)
			impact = "Package will be uninstalled for primary user but remains in system image."
			undoable = true
		} else {
			title = fmt.Sprintf("Uninstall application %s", pkg)
			impact = "Application and internal app data will be removed. Backed up APK can be restored."
			undoable = true
		}
	case ActionRevokePermission:
		title = fmt.Sprintf("Revoke permission %s from %s", perm, pkg)
		impact = fmt.Sprintf("Application can no longer access %s.", perm)
		undoable = true
	case ActionClearAppop:
		title = fmt.Sprintf("Clear app operation %s for %s", appop, pkg)
		impact = fmt.Sprintf("Operation mode reset to default/denied.")
		undoable = true
	default:
		actionType = ActionDisable
		title = fmt.Sprintf("Disable package %s", pkg)
		impact = "Package will be frozen."
	}

	act := Action{
		ID:               fmt.Sprintf("act-%d", time.Now().UnixNano()),
		Type:             actionType,
		Package:          pkg,
		Title:            title,
		Impact:           impact,
		Undoable:         undoable,
		IsBlocked:        isBlocked,
		TargetPermission: perm,
		TargetAppop:      appop,
		IsSystemApp:      isSystem,
	}

	if isBlocked {
		act.BlockedReason = &reason
	}

	return act
}

// CreatePlan groups actions into an executable plan with backup determination.
func (m *Manager) CreatePlan(actions ...Action) Plan {
	requiresBackup := false
	for _, a := range actions {
		if a.Type == ActionUninstall && !a.IsSystemApp {
			requiresBackup = true
			break
		}
	}
	return Plan{
		Actions:        actions,
		RequiresBackup: requiresBackup,
	}
}

func (m *Manager) logFailedAction(serial string, action Action, err error) *ActionLogEntry {
	entryID := fmt.Sprintf("log-%d", time.Now().UnixNano())
	now := time.Now().UTC()
	errStr := err.Error()

	entry := &ActionLogEntry{
		ID:        entryID,
		Timestamp: now.Format(time.RFC3339),
		Action:    action,
		Status:    "failed",
		Error:     &errStr,
	}

	if m.store != nil {
		_ = m.store.LogAction(store.ActionLogRecord{
			ID:           entryID,
			DeviceSerial: serial,
			Timestamp:    now,
			ActionType:   action.Type,
			Package:      action.Package,
			Details: map[string]any{
				"action": action,
				"error":  errStr,
			},
			Reversible: false,
		})
	}

	return entry
}

// Execute performs the remediation action with safety guardrails and audit logging.
func (m *Manager) Execute(ctx context.Context, serial string, action Action) (*ActionLogEntry, error) {
	if action.IsBlocked {
		reason := "Blocked by guardrail"
		if action.BlockedReason != nil {
			reason = *action.BlockedReason
		}
		err := fmt.Errorf(`JERION_ERR:{"code":"ActionBlocked","message":"%s","remedy":"Protected critical system packages cannot be removed or disabled."}`, reason)
		return m.logFailedAction(serial, action, err), err
	}

	if err := ValidatePackageName(action.Package); err != nil {
		wErr := fmt.Errorf(`JERION_ERR:{"code":"ParseFailure","message":"%s","remedy":"Ensure valid package name format"}`, err.Error())
		return m.logFailedAction(serial, action, wErr), wErr
	}

	var backupPath *string

	// 1. If uninstalling a user package, backup APK first
	if action.Type == ActionUninstall && !action.IsSystemApp {
		_ = os.MkdirAll(m.backupDir, 0755)
		puller := inventory.NewPuller(m.runner, m.logger)
		basePath, _, _, err := puller.Pull(ctx, serial, action.Package, m.backupDir)
		if err != nil || basePath == "" {
			m.logger.Error("remediate: APK backup failed before uninstall", "package", action.Package, "err", err)
			wErr := fmt.Errorf(`JERION_ERR:{"code":"BackupFailed","message":"Failed to backup APK for %s before uninstall: %v","remedy":"Device refused APK extraction or storage error. Cannot uninstall safely without backup."}`, action.Package, err)
			return m.logFailedAction(serial, action, wErr), wErr
		}
		backupPath = &basePath
		m.logger.Info("remediate: APK backed up before uninstall", "package", action.Package, "path", basePath)
	}

	// 2. Execute command via ADB
	var execErr error
	switch action.Type {
	case ActionDisable:
		_, stderr, err := m.runner.Run(ctx, serial, "shell", "pm", "disable-user", "--user", "0", action.Package)
		if err != nil {
			execErr = fmt.Errorf("disable failed: %v (%s)", err, string(stderr))
		}
	case ActionUninstall:
		if action.IsSystemApp {
			_, stderr, err := m.runner.Run(ctx, serial, "shell", "pm", "uninstall", "--user", "0", action.Package)
			if err != nil {
				execErr = fmt.Errorf("system uninstall failed: %v (%s)", err, string(stderr))
			}
		} else {
			_, stderr, err := m.runner.Run(ctx, serial, "shell", "pm", "uninstall", action.Package)
			if err != nil {
				execErr = fmt.Errorf("uninstall failed: %v (%s)", err, string(stderr))
			}
		}
	case ActionRevokePermission:
		if action.TargetPermission == "" {
			err := fmt.Errorf(`JERION_ERR:{"code":"InvalidAction","message":"Target permission cannot be empty for revoke_permission action","remedy":"Specify a valid permission name"}`)
			return m.logFailedAction(serial, action, err), err
		}
		if err := ValidatePermissionName(action.TargetPermission); err != nil {
			wErr := fmt.Errorf(`JERION_ERR:{"code":"ParseFailure","message":"%s","remedy":"Verify permission format"}`, err.Error())
			return m.logFailedAction(serial, action, wErr), wErr
		}
		_, stderr, err := m.runner.Run(ctx, serial, "shell", "pm", "revoke", action.Package, action.TargetPermission)
		if err != nil {
			execErr = fmt.Errorf("revoke permission failed: %v (%s)", err, string(stderr))
		}
	case ActionClearAppop:
		if action.TargetAppop == "" {
			err := fmt.Errorf(`JERION_ERR:{"code":"InvalidAction","message":"Target appop cannot be empty for clear_appop action","remedy":"Specify a valid appop name"}`)
			return m.logFailedAction(serial, action, err), err
		}
		_, stderr, err := m.runner.Run(ctx, serial, "shell", "cmd", "appops", "set", action.Package, action.TargetAppop, "default")
		if err != nil {
			// Fallback to appops without cmd
			_, stderr2, err2 := m.runner.Run(ctx, serial, "shell", "appops", "set", action.Package, action.TargetAppop, "default")
			if err2 != nil {
				execErr = fmt.Errorf("clear appop failed: %v (%s / %s)", err, string(stderr), string(stderr2))
			}
		}
	default:
		execErr = fmt.Errorf("unrecognized remediation action type: %s", action.Type)
	}

	// 3. Prepare audit log entry
	status := "success"
	var errStr *string
	if execErr != nil {
		status = "failed"
		s := execErr.Error()
		errStr = &s
	}

	entryID := fmt.Sprintf("log-%d", time.Now().UnixNano())
	now := time.Now().UTC()

	var bpVal string
	if backupPath != nil {
		bpVal = *backupPath
	}

	details := map[string]any{
		"action":           action,
		"backupPath":       bpVal,
		"targetPermission": action.TargetPermission,
		"targetAppop":      action.TargetAppop,
		"isSystemApp":      action.IsSystemApp,
	}

	if m.store != nil {
		_ = m.store.LogAction(store.ActionLogRecord{
			ID:           entryID,
			DeviceSerial: serial,
			Timestamp:    now,
			ActionType:   action.Type,
			Package:      action.Package,
			Details:      details,
			Reversible:   action.Undoable,
			Undone:       false,
		})
	}

	logEntry := &ActionLogEntry{
		ID:         entryID,
		Timestamp:  now.Format(time.RFC3339),
		Action:     action,
		Status:     status,
		Error:      errStr,
		BackupPath: backupPath,
	}

	if execErr != nil {
		return logEntry, execErr
	}

	return logEntry, nil
}

// Undo reverses a previously executed remediation action.
func (m *Manager) Undo(ctx context.Context, serial string, logEntryID string) (*ActionLogEntry, error) {
	if m.store == nil {
		return nil, fmt.Errorf("database store not configured")
	}

	record, err := m.store.GetAction(logEntryID)
	if err != nil {
		return nil, fmt.Errorf("action not found: %w", err)
	}

	if record.Undone {
		return nil, fmt.Errorf("action %s is already undone", logEntryID)
	}

	pkg := record.Package
	var undoErr error

	switch record.ActionType {
	case ActionDisable:
		_, stderr, err := m.runner.Run(ctx, serial, "shell", "pm", "enable", pkg)
		if err != nil {
			undoErr = fmt.Errorf("enable package failed: %v (%s)", err, string(stderr))
		}
	case ActionUninstall:
		isSystem := false
		if isSys, ok := record.Details["isSystemApp"].(bool); ok && isSys {
			isSystem = true
		}

		if isSystem {
			_, stderr, err := m.runner.Run(ctx, serial, "shell", "cmd", "package", "install-existing", pkg)
			if err != nil {
				undoErr = fmt.Errorf("install-existing system app failed: %v (%s)", err, string(stderr))
			}
		} else {
			// Restore from backup APK if available
			var bPath string
			if bp, ok := record.Details["backupPath"].(string); ok && bp != "" {
				bPath = bp
			} else if bpPtr, ok := record.Details["backupPath"].(*string); ok && bpPtr != nil {
				bPath = *bpPtr
			}

			if bPath != "" && fileExists(bPath) {
				_, stderr, err := m.runner.Run(ctx, serial, "install", "-r", bPath)
				if err != nil {
					undoErr = fmt.Errorf("reinstall from backup APK failed: %v (%s)", err, string(stderr))
				}
			} else {
				undoErr = fmt.Errorf("cannot undo uninstall: backup APK is missing or not pulled")
			}
		}
	case ActionRevokePermission:
		perm := ""
		if p, ok := record.Details["targetPermission"].(string); ok {
			perm = p
		}
		if perm != "" {
			_, stderr, err := m.runner.Run(ctx, serial, "shell", "pm", "grant", pkg, perm)
			if err != nil {
				undoErr = fmt.Errorf("grant permission failed: %v (%s)", err, string(stderr))
			}
		}
	case ActionClearAppop:
		appop := ""
		if op, ok := record.Details["targetAppop"].(string); ok {
			appop = op
		}
		if appop != "" {
			_, stderr, err := m.runner.Run(ctx, serial, "shell", "cmd", "appops", "set", pkg, appop, "allow")
			if err != nil {
				undoErr = fmt.Errorf("allow appop failed: %v (%s)", err, string(stderr))
			}
		}
	default:
		undoErr = fmt.Errorf("unsupported undo action type: %s", record.ActionType)
	}

	if undoErr != nil {
		return nil, undoErr
	}

	if err := m.store.MarkActionUndone(logEntryID); err != nil {
		return nil, fmt.Errorf("mark undone in database: %w", err)
	}

	record.Undone = true
	now := time.Now().UTC()
	record.UndoneAt = &now

	// Construct returned ActionLogEntry
	return &ActionLogEntry{
		ID:        record.ID,
		Timestamp: record.Timestamp.Format(time.RFC3339),
		Action: Action{
			ID:       record.ID,
			Type:     record.ActionType,
			Package:  record.Package,
			Title:    fmt.Sprintf("Undo %s on %s", record.ActionType, record.Package),
			Undoable: false,
		},
		Status: "undone",
	}, nil
}

// GetHistory retrieves action history from database.
func (m *Manager) GetHistory(ctx context.Context, serial string) ([]ActionLogEntry, error) {
	if m.store == nil {
		return []ActionLogEntry{}, nil
	}

	records, err := m.store.ListActionLogs(serial)
	if err != nil {
		return nil, err
	}

	var res []ActionLogEntry
	for _, r := range records {
		status := "success"
		if r.Undone {
			status = "undone"
		}

		var bPath *string
		if bp, ok := r.Details["backupPath"].(string); ok && bp != "" {
			bPath = &bp
		}

		res = append(res, ActionLogEntry{
			ID:        r.ID,
			Timestamp: r.Timestamp.Format(time.RFC3339),
			Action: Action{
				ID:       r.ID,
				Type:     r.ActionType,
				Package:  r.Package,
				Title:    fmt.Sprintf("%s on %s", r.ActionType, r.Package),
				Undoable: r.Reversible && !r.Undone,
			},
			Status:     status,
			BackupPath: bPath,
		})
	}

	return res, nil
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
