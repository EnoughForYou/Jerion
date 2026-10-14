package store

import (
	"errors"
	"testing"
	"time"
)

func setupTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := OpenMemory()
	if err != nil {
		t.Fatalf("failed to open memory store: %v", err)
	}
	t.Cleanup(func() {
		_ = s.Close()
	})
	return s
}

func TestStoreMigrations(t *testing.T) {
	s := setupTestStore(t)

	var count int
	err := s.DB().QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query schema_migrations: %v", err)
	}
	if count < 1 {
		t.Fatalf("expected at least 1 migration applied, got %d", count)
	}
}

func TestStoreSettings(t *testing.T) {
	s := setupTestStore(t)

	// Initial get
	_, err := s.GetSetting("theme")
	if !errors.Is(err, ErrSettingNotFound) {
		t.Fatalf("expected ErrSettingNotFound, got %v", err)
	}

	// Set & Get
	if err := s.SetSetting("theme", "dark"); err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}
	val, err := s.GetSetting("theme")
	if err != nil || val != "dark" {
		t.Fatalf("unexpected setting val=%s, err=%v", val, err)
	}

	// GetAll
	_ = s.SetSetting("adb_path", "/usr/bin/adb")
	all, err := s.GetAllSettings()
	if err != nil || len(all) != 2 {
		t.Fatalf("GetAllSettings returned %v, err=%v", all, err)
	}

	// Delete
	if err := s.DeleteSetting("theme"); err != nil {
		t.Fatalf("DeleteSetting failed: %v", err)
	}
	_, err = s.GetSetting("theme")
	if !errors.Is(err, ErrSettingNotFound) {
		t.Fatalf("expected ErrSettingNotFound after delete, got %v", err)
	}
}

func TestStoreDevices(t *testing.T) {
	s := setupTestStore(t)

	dev := DeviceRecord{
		Serial:         "TEST-DEVICE-01",
		Model:          "Infinix Note 11S",
		Manufacturer:   "Infinix",
		AndroidVersion: "12",
		SDK:            31,
		SecurityPatch:  "2022-04-05",
		Fingerprint:    "Infinix/X698/Infinix-X698:12/SP1A.210812.016/220415V437:user/release-keys",
	}

	if err := s.UpsertDevice(dev); err != nil {
		t.Fatalf("UpsertDevice failed: %v", err)
	}

	retrieved, err := s.GetDevice("TEST-DEVICE-01")
	if err != nil {
		t.Fatalf("GetDevice failed: %v", err)
	}
	if retrieved.Model != "Infinix Note 11S" || retrieved.SDK != 31 {
		t.Fatalf("unexpected device record: %+v", retrieved)
	}

	list, err := s.ListDevices()
	if err != nil || len(list) != 1 {
		t.Fatalf("ListDevices failed: len=%d, err=%v", len(list), err)
	}
}

func TestStoreScansAndFindings(t *testing.T) {
	s := setupTestStore(t)

	now := time.Now().UTC()
	scan := ScanRecord{
		ID:           "scan-001",
		DeviceSerial: "TEST-DEVICE-01",
		Mode:         "quick",
		StartedAt:    now.Add(-10 * time.Second),
		FinishedAt:   now,
		RulesVersion: "1.0.0",
		AppVersion:   "0.4.0",
		Partial:      false,
		TotalApps:    42,
		Findings: []FindingRecord{
			{
				ID:                "f-001",
				ScanID:            "scan-001",
				RuleID:            "ADWARE-001",
				Package:           "com.adware.cleaner",
				Category:          "adware",
				Severity:          "high",
				Confidence:        "high",
				Weight:            80,
				Title:             "Aggressive Fullscreen Adware",
				WhyFlagged:        "Manifest declares TYPE_APPLICATION_OVERLAY without justification",
				FalsePositiveHint: "System apps rarely use this permission",
				Evidence: []EvidenceRecord{
					{Source: "manifest", Snippet: "<uses-permission name=\"SYSTEM_ALERT_WINDOW\"/>"},
				},
				Actions: []string{"uninstall", "disable"},
			},
		},
	}

	if err := s.SaveScanResult(scan); err != nil {
		t.Fatalf("SaveScanResult failed: %v", err)
	}

	// Retrieve
	loaded, err := s.GetScanResult("scan-001")
	if err != nil {
		t.Fatalf("GetScanResult failed: %v", err)
	}
	if loaded.Mode != "quick" || len(loaded.Findings) != 1 || loaded.TotalApps != 42 {
		t.Fatalf("unexpected loaded scan: %+v", loaded)
	}
	if loaded.Findings[0].RuleID != "ADWARE-001" || len(loaded.Findings[0].Evidence) != 1 {
		t.Fatalf("unexpected finding: %+v", loaded.Findings[0])
	}

	// Summary list
	summaries, err := s.ListScanSummaries("TEST-DEVICE-01")
	if err != nil || len(summaries) != 1 {
		t.Fatalf("ListScanSummaries failed: %v", err)
	}
	if summaries[0].FindingCount != 1 || summaries[0].HighSeverityCount != 1 || summaries[0].TotalApps != 42 {
		t.Fatalf("unexpected summary count: %+v", summaries[0])
	}

	// Delete
	if err := s.DeleteScanResult("scan-001"); err != nil {
		t.Fatalf("DeleteScanResult failed: %v", err)
	}
	_, err = s.GetScanResult("scan-001")
	if !errors.Is(err, ErrScanNotFound) {
		t.Fatalf("expected ErrScanNotFound after delete, got %v", err)
	}
}

func TestStoreActionsLog(t *testing.T) {
	s := setupTestStore(t)

	act := ActionLogRecord{
		ID:           "act-001",
		DeviceSerial: "TEST-DEVICE-01",
		Timestamp:    time.Now().UTC(),
		ActionType:   "disable_package",
		Package:      "com.suspicious.app",
		Details:      map[string]any{"backup": true, "path": "/work/backup.apk"},
		Reversible:   true,
		Undone:       false,
	}

	if err := s.LogAction(act); err != nil {
		t.Fatalf("LogAction failed: %v", err)
	}

	retrieved, err := s.GetAction("act-001")
	if err != nil {
		t.Fatalf("GetAction failed: %v", err)
	}
	if !retrieved.Reversible || retrieved.Undone {
		t.Fatalf("unexpected action record: %+v", retrieved)
	}

	// Mark Undone
	if err := s.MarkActionUndone("act-001"); err != nil {
		t.Fatalf("MarkActionUndone failed: %v", err)
	}

	retrievedAgain, _ := s.GetAction("act-001")
	if !retrievedAgain.Undone || retrievedAgain.UndoneAt == nil {
		t.Fatalf("expected Undone=true with UndoneAt set, got %+v", retrievedAgain)
	}

	list, err := s.ListActionLogs("TEST-DEVICE-01")
	if err != nil || len(list) != 1 {
		t.Fatalf("ListActionLogs failed: %v", err)
	}
}

func TestStoreAllowlist(t *testing.T) {
	s := setupTestStore(t)

	entry := AllowlistRecord{
		ID:         "allow-001",
		Package:    "com.google.android.youtube",
		CertSHA256: "AFB1234567890ABCDEF",
		Reason:     "Official YouTube client",
		AddedAt:    time.Now().UTC(),
	}

	if err := s.AddAllowlistEntry(entry); err != nil {
		t.Fatalf("AddAllowlistEntry failed: %v", err)
	}

	ok, err := s.IsAllowlisted("com.google.android.youtube", "AFB1234567890ABCDEF")
	if err != nil || !ok {
		t.Fatalf("IsAllowlisted expected true, got ok=%v, err=%v", ok, err)
	}

	ok2, _ := s.IsAllowlisted("com.evil.app", "ANY")
	if ok2 {
		t.Fatalf("IsAllowlisted for unknown app expected false")
	}

	list, err := s.ListAllowlist()
	if err != nil || len(list) != 1 {
		t.Fatalf("ListAllowlist failed: %v", err)
	}

	if err := s.RemoveAllowlistEntry("allow-001"); err != nil {
		t.Fatalf("RemoveAllowlistEntry failed: %v", err)
	}
	okAfterRemove, _ := s.IsAllowlisted("com.google.android.youtube", "AFB1234567890ABCDEF")
	if okAfterRemove {
		t.Fatalf("expected false after remove")
	}
}

func TestStoreIntelHashes(t *testing.T) {
	s := setupTestStore(t)

	hash := IntelHashRecord{
		SHA256:         "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		Classification: "adware",
		ThreatName:     "Android.Adware.MobiDash",
		Source:         "internal_rules",
	}

	if err := s.UpsertIntelHash(hash); err != nil {
		t.Fatalf("UpsertIntelHash failed: %v", err)
	}

	// Lookup (case insensitive check)
	retrieved, err := s.LookupIntelHash("E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855")
	if err != nil {
		t.Fatalf("LookupIntelHash failed: %v", err)
	}
	if retrieved.Classification != "adware" || retrieved.ThreatName != "Android.Adware.MobiDash" {
		t.Fatalf("unexpected intel record: %+v", retrieved)
	}

	// Bulk import
	bulk := []IntelHashRecord{
		{SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Classification: "clean", Source: "user"},
		{SHA256: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Classification: "malware", ThreatName: "Trojan.Dropper", Source: "user"},
	}
	if err := s.ImportIntelHashes(bulk); err != nil {
		t.Fatalf("ImportIntelHashes failed: %v", err)
	}

	h2, err := s.LookupIntelHash("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	if err != nil || h2.Classification != "malware" {
		t.Fatalf("unexpected h2 record: %+v, err=%v", h2, err)
	}
}
