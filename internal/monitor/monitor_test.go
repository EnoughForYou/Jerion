package monitor_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/EnoughForYou/Jerion/internal/adb"
	"github.com/EnoughForYou/Jerion/internal/logx"
	"github.com/EnoughForYou/Jerion/internal/monitor"
)

func TestMonitorManagerLifecycleAndRecord(t *testing.T) {
	fake := adb.NewFakeRunner()
	logger := logx.GetDefault()
	mgr := monitor.NewManager(fake, logger)

	// Status before start
	status := mgr.Status()
	if status.IsActive {
		t.Fatalf("expected inactive status before start")
	}

	// Register event and candidate callbacks
	var eventCount int32
	var candidateCount int32
	mgr.OnEvent(func(evt monitor.MonitorEvent) {
		atomic.AddInt32(&eventCount, 1)
	})
	mgr.OnCandidate(func(cand monitor.MonitorCandidate) {
		atomic.AddInt32(&candidateCount, 1)
	})

	// Record a window focus event
	evt1 := monitor.MonitorEvent{
		ID:        "evt-1",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Type:      monitor.EventWindowFocusChange,
		Package:   "com.adware.popupapp",
		Details: map[string]any{
			"activity": ".MainActivity",
		},
	}
	mgr.RecordEvent(evt1)

	if atomic.LoadInt32(&eventCount) != 1 {
		t.Fatalf("expected 1 event, got %d", atomic.LoadInt32(&eventCount))
	}
	if atomic.LoadInt32(&candidateCount) != 1 {
		t.Fatalf("expected 1 candidate, got %d", atomic.LoadInt32(&candidateCount))
	}

	snap := mgr.Snapshot()
	if len(snap.Events) != 1 {
		t.Fatalf("expected 1 event in snapshot, got %d", len(snap.Events))
	}
	if len(snap.Candidates) != 1 {
		t.Fatalf("expected 1 candidate in snapshot, got %d", len(snap.Candidates))
	}
	if snap.Candidates[0].Package != "com.adware.popupapp" {
		t.Fatalf("unexpected candidate package: %s", snap.Candidates[0].Package)
	}

	// Record overlay shown event
	evt2 := monitor.MonitorEvent{
		ID:        "evt-2",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Type:      monitor.EventOverlayShown,
		Package:   "com.adware.popupapp",
		Details: map[string]any{
			"windowType": "TYPE_APPLICATION_OVERLAY",
		},
	}
	mgr.RecordEvent(evt2)

	snap2 := mgr.Snapshot()
	if snap2.Candidates[0].Score < 80 {
		t.Fatalf("expected elevated score for overlay candidate, got %d", snap2.Candidates[0].Score)
	}

	// Mark user popup
	mgr.MarkUserPopup(time.Now())
	snap3 := mgr.Snapshot()
	if snap3.Candidates[0].Score < 90 {
		t.Fatalf("expected score >= 90 after MarkUserPopup, got %d", snap3.Candidates[0].Score)
	}
}

func TestMonitorCollectorLoop(t *testing.T) {
	fake := adb.NewFakeRunner()
	fake.SetResponse("dumpsys window", `
  mCurrentFocus=Window{12345 u0 com.popup.spammer/com.popup.spammer.AdActivity}
  Window #1 Window{67890 u0 com.popup.spammer/com.popup.spammer.OverlayView}:
    ty=TYPE_APPLICATION_OVERLAY
`, "", nil)

	logger := logx.GetDefault()
	mgr := monitor.NewManager(fake, logger)

	var recordedEvents []monitor.MonitorEvent
	mgr.OnEvent(func(evt monitor.MonitorEvent) {
		recordedEvents = append(recordedEvents, evt)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := mgr.Start(ctx, "device-1")
	if err != nil {
		t.Fatalf("failed to start monitor: %v", err)
	}

	status := mgr.Status()
	if !status.IsActive {
		t.Fatalf("expected active status")
	}

	// Let the ticker run for 1.2s to capture telemetry
	time.Sleep(1200 * time.Millisecond)

	mgr.Stop()
	statusAfter := mgr.Status()
	if statusAfter.IsActive {
		t.Fatalf("expected inactive status after Stop")
	}

	snap := mgr.Snapshot()
	if len(snap.Candidates) == 0 {
		t.Fatalf("expected candidates detected by collector loop, got 0")
	}
	if snap.Candidates[0].Package != "com.popup.spammer" {
		t.Fatalf("expected candidate com.popup.spammer, got %s", snap.Candidates[0].Package)
	}
}
