package secrets

import (
	"errors"
	"testing"
)

func TestMemoryKeyring(t *testing.T) {
	kr := NewMemoryKeyring()

	// Initial get returns ErrNotFound
	_, err := kr.Get("test", "user1")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	// Empty secret rejected
	if err := kr.Set("test", "user1", ""); !errors.Is(err, ErrInvalidSecret) {
		t.Fatalf("expected ErrInvalidSecret, got %v", err)
	}

	// Set and Get
	if err := kr.Set("test", "user1", "secret123"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := kr.Get("test", "user1")
	if err != nil || val != "secret123" {
		t.Fatalf("Get returned unexpected: val=%s, err=%v", val, err)
	}

	// Delete
	if err := kr.Delete("test", "user1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = kr.Get("test", "user1")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after Delete, got %v", err)
	}
}

func TestManagerOperations(t *testing.T) {
	kr := NewMemoryKeyring()
	mgr := NewManager(kr)

	// Probe test
	if err := mgr.ProbeTest(); err != nil {
		t.Fatalf("ProbeTest failed: %v", err)
	}

	// VirusTotal key flow
	if mgr.HasVirusTotalKey() {
		t.Fatalf("expected HasVirusTotalKey false initially")
	}

	if err := mgr.SaveVirusTotalKey("vt-secret-api-key"); err != nil {
		t.Fatalf("SaveVirusTotalKey failed: %v", err)
	}

	if !mgr.HasVirusTotalKey() {
		t.Fatalf("expected HasVirusTotalKey true after save")
	}

	key, err := mgr.GetVirusTotalKey()
	if err != nil || key != "vt-secret-api-key" {
		t.Fatalf("GetVirusTotalKey returned %s, err: %v", key, err)
	}

	if err := mgr.DeleteVirusTotalKey(); err != nil {
		t.Fatalf("DeleteVirusTotalKey failed: %v", err)
	}

	if mgr.HasVirusTotalKey() {
		t.Fatalf("expected HasVirusTotalKey false after delete")
	}
}

func TestOSKeyringInstantiation(t *testing.T) {
	osKr := NewOSKeyring()
	if osKr == nil {
		t.Fatal("expected NewOSKeyring not nil")
	}
}
