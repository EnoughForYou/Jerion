package remediate_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EnoughForYou/Jerion/internal/adb"
	"github.com/EnoughForYou/Jerion/internal/logx"
	"github.com/EnoughForYou/Jerion/internal/remediate"
	"github.com/EnoughForYou/Jerion/internal/store"
)

func TestGuardrails(t *testing.T) {
	// 1. Critical system package
	blocked, reason := remediate.CheckGuardrail("com.android.systemui")
	if !blocked {
		t.Fatalf("expected com.android.systemui to be blocked")
	}
	if reason == "" {
		t.Fatalf("expected reason for blocked package")
	}

	// 2. Core android
	blockedAndroid, _ := remediate.CheckGuardrail("android")
	if !blockedAndroid {
		t.Fatalf("expected core android to be blocked")
	}

	// 3. Transsion XOS system app
	blockedXOS, _ := remediate.CheckGuardrail("com.transsion.launcher")
	if !blockedXOS {
		t.Fatalf("expected com.transsion.launcher to be blocked by safety guardrail")
	}

	// 4. User app
	blockedUser, _ := remediate.CheckGuardrail("com.suspicious.adware")
	if blockedUser {
		t.Fatalf("expected com.suspicious.adware to NOT be blocked")
	}

	// 5. Comprehensive shell injection attempts
	maliciousPackages := []string{
		"com.foo.bar; rm -rf /",
		"com.foo`id`",
		"com.foo$(whoami)",
		"com.foo|cat /etc/passwd",
		"com.foo&calc.exe",
		"com.foo\nrm -rf /",
		"com.foo>out.txt",
		"com.foo<in.txt",
	}
	for _, pkg := range maliciousPackages {
		if err := remediate.ValidatePackageName(pkg); err == nil {
			t.Fatalf("expected error for malicious package name: %q", pkg)
		}
	}

	maliciousPermissions := []string{
		"android.permission.INTERNET && whoami",
		"android.permission.CAMERA; id",
		"android.permission.RECORD_AUDIO | rm -rf",
		"android.permission.WAKE_LOCK`date`",
	}
	for _, perm := range maliciousPermissions {
		if err := remediate.ValidatePermissionName(perm); err == nil {
			t.Fatalf("expected error for malicious permission name: %q", perm)
		}
	}
}

func TestRemediationExecuteAndUndo(t *testing.T) {
	fake := adb.NewFakeRunner()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_remediate.db")
	st, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}
	defer st.Close()

	logger := logx.GetDefault()
	backupDir := filepath.Join(tmpDir, "backups")
	mgr := remediate.NewManager(fake, st, logger, backupDir)

	ctx := context.Background()
	serial := "test-device"

	// 1. Attempt blocked action on SystemUI
	blockedAct := mgr.CreateAction(remediate.ActionDisable, "com.android.systemui", "", "", true)
	if !blockedAct.IsBlocked {
		t.Fatalf("expected action to be marked blocked")
	}
	_, err = mgr.Execute(ctx, serial, blockedAct)
	if err == nil || !strings.Contains(err.Error(), "ActionBlocked") {
		t.Fatalf("expected ActionBlocked error, got: %v", err)
	}

	// 2. Disable user package and Undo
	disableAct := mgr.CreateAction(remediate.ActionDisable, "com.adware.spam", "", "", false)
	fake.SetResponse("pm disable-user --user 0 com.adware.spam", "Package com.adware.spam new state: disabled-user", "", nil)
	fake.SetResponse("pm enable com.adware.spam", "Package com.adware.spam new state: enabled", "", nil)

	logEntry, err := mgr.Execute(ctx, serial, disableAct)
	if err != nil {
		t.Fatalf("execute disable failed: %v", err)
	}
	if logEntry.Status != "success" {
		t.Fatalf("expected status success, got %s", logEntry.Status)
	}

	// Verify history
	history, err := mgr.GetHistory(ctx, serial)
	if err != nil {
		t.Fatalf("get history failed: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(history))
	}

	// Undo disable
	undoneEntry, err := mgr.Undo(ctx, serial, logEntry.ID)
	if err != nil {
		t.Fatalf("undo disable failed: %v", err)
	}
	if undoneEntry.Status != "undone" {
		t.Fatalf("expected undone status, got %s", undoneEntry.Status)
	}

	// 3. System app uninstall and Undo
	sysAct := mgr.CreateAction(remediate.ActionUninstall, "com.oem.bloatware", "", "", true)
	fake.SetResponse("pm uninstall --user 0 com.oem.bloatware", "Success", "", nil)
	fake.SetResponse("cmd package install-existing com.oem.bloatware", "Package com.oem.bloatware installed", "", nil)

	sysLog, err := mgr.Execute(ctx, serial, sysAct)
	if err != nil {
		t.Fatalf("execute sys uninstall failed: %v", err)
	}

	undoneSys, err := mgr.Undo(ctx, serial, sysLog.ID)
	if err != nil {
		t.Fatalf("undo sys uninstall failed: %v", err)
	}
	if undoneSys.Status != "undone" {
		t.Fatalf("expected undone status for sys app, got %s", undoneSys.Status)
	}

	// 4. Revoke permission and Undo
	permAct := mgr.CreateAction(remediate.ActionRevokePermission, "com.adware.spam", "android.permission.SYSTEM_ALERT_WINDOW", "", false)
	fake.SetResponse("pm revoke com.adware.spam android.permission.SYSTEM_ALERT_WINDOW", "", "", nil)
	fake.SetResponse("pm grant com.adware.spam android.permission.SYSTEM_ALERT_WINDOW", "", "", nil)

	permLog, err := mgr.Execute(ctx, serial, permAct)
	if err != nil {
		t.Fatalf("execute perm revoke failed: %v", err)
	}

	undonePerm, err := mgr.Undo(ctx, serial, permLog.ID)
	if err != nil {
		t.Fatalf("undo perm revoke failed: %v", err)
	}
	if undonePerm.Status != "undone" {
		t.Fatalf("expected undone status for permission, got %s", undonePerm.Status)
	}

	// 5. User app uninstall fails when APK backup fails (Guardrail enforcement)
	failUserAct := mgr.CreateAction(remediate.ActionUninstall, "com.malware.nobackup", "", "", false)
	// pm path fails / returns error
	fake.SetResponse("pm path com.malware.nobackup", "", "package not found", os.ErrNotExist)
	_, err = mgr.Execute(ctx, serial, failUserAct)
	if err == nil || !strings.Contains(err.Error(), "BackupFailed") {
		t.Fatalf("expected BackupFailed error when APK backup fails, got: %v", err)
	}

	// 6. User app uninstall succeeds when APK backup succeeds, and restores on Undo
	succUserAct := mgr.CreateAction(remediate.ActionUninstall, "com.malware.backedup", "", "", false)
	fake.SetResponse("pm path com.malware.backedup", "package:/data/app/~~dummy/base.apk\n", "", nil)
	fake.SetHandler(
		func(devSerial string, args []string) bool {
			return len(args) >= 3 && args[0] == "pull"
		},
		func(devSerial string, args []string) ([]byte, []byte, error) {
			dest := args[2]
			_ = os.WriteFile(dest, []byte("dummy-backup-apk"), 0o644)
			return []byte("1 file pulled"), nil, nil
		},
	)
	fake.SetResponse("pm uninstall com.malware.backedup", "Success", "", nil)
	fake.SetHandler(
		func(devSerial string, args []string) bool {
			return len(args) >= 3 && args[0] == "install" && args[1] == "-r"
		},
		func(devSerial string, args []string) ([]byte, []byte, error) {
			return []byte("Success"), nil, nil
		},
	)

	userLog, err := mgr.Execute(ctx, serial, succUserAct)
	if err != nil {
		t.Fatalf("execute user uninstall with backup failed: %v", err)
	}
	if userLog.Status != "success" {
		t.Fatalf("expected status success, got %s", userLog.Status)
	}
	if userLog.BackupPath == nil || *userLog.BackupPath == "" {
		t.Fatalf("expected non-empty backupPath in user uninstall log")
	}

	undoneUser, err := mgr.Undo(ctx, serial, userLog.ID)
	if err != nil {
		t.Fatalf("undo user uninstall failed: %v", err)
	}
	if undoneUser.Status != "undone" {
		t.Fatalf("expected undone status for restored user app, got %s", undoneUser.Status)
	}
}

func TestRemediationPlanCreation(t *testing.T) {
	mgr := remediate.NewManager(nil, nil, nil, "")
	act1 := mgr.CreateAction(remediate.ActionDisable, "com.test.one", "", "", false)
	act2 := mgr.CreateAction(remediate.ActionUninstall, "com.test.two", "", "", false)

	plan := mgr.CreatePlan(act1, act2)
	if len(plan.Actions) != 2 {
		t.Fatalf("expected 2 actions in plan, got %d", len(plan.Actions))
	}
	if !plan.RequiresBackup {
		t.Fatalf("expected RequiresBackup to be true because of non-system uninstall")
	}
}

func TestRemediateEmptyTargetRejection(t *testing.T) {
	fake := adb.NewFakeRunner()
	mgr := remediate.NewManager(fake, nil, logx.GetDefault(), "")
	ctx := context.Background()

	// 1. Revoke permission with empty target must fail
	emptyPermAct := remediate.Action{
		ID:      "act-empty-perm",
		Type:    remediate.ActionRevokePermission,
		Package: "com.test.app",
	}
	_, err := mgr.Execute(ctx, "dev-1", emptyPermAct)
	if err == nil || !strings.Contains(err.Error(), "InvalidAction") {
		t.Fatalf("expected InvalidAction error for empty revoke permission target, got: %v", err)
	}

	// 2. Clear appop with empty target must fail
	emptyAppopAct := remediate.Action{
		ID:      "act-empty-appop",
		Type:    remediate.ActionClearAppop,
		Package: "com.test.app",
	}
	_, err = mgr.Execute(ctx, "dev-1", emptyAppopAct)
	if err == nil || !strings.Contains(err.Error(), "InvalidAction") {
		t.Fatalf("expected InvalidAction error for empty clear appop target, got: %v", err)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
