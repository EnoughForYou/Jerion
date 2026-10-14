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

func TestMonitorConcurrentSnapshotNoDeadlock(t *testing.T) {
	fake := adb.NewFakeRunner()
	logger := logx.GetDefault()
	mgr := monitor.NewManager(fake, logger)

	done := make(chan struct{})
	go func() {
		// Run readers and writers concurrently for 200 iterations
		for i := 0; i < 200; i++ {
			go func(iter int) {
				mgr.RecordEvent(monitor.MonitorEvent{
					ID:        "evt-concurrent",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					Type:      monitor.EventForegroundSwitch,
					Package:   "com.concurrent.app",
					Details:   map[string]any{"i": iter},
				})
			}(i)

			go func() {
				_ = mgr.Snapshot()
			}()

			go func() {
				mgr.MarkUserPopup(time.Now().UTC())
			}()

			go func() {
				_ = mgr.Status()
			}()
		}
		time.Sleep(100 * time.Millisecond)
		close(done)
	}()

	select {
	case <-done:
		// success, no deadlock
	case <-time.After(3 * time.Second):
		t.Fatalf("deadlock detected during concurrent Snapshot and writer operations")
	}
}

func TestMarkUserPopupThreeSecondWindow(t *testing.T) {
	fake := adb.NewFakeRunner()
	logger := logx.GetDefault()
	mgr := monitor.NewManager(fake, logger)

	now := time.Now().UTC()

	// 1. Record an event that happened 10 seconds ago
	oldTime := now.Add(-10 * time.Second).Format(time.RFC3339)
	mgr.RecordEvent(monitor.MonitorEvent{
		ID:        "evt-old",
		Timestamp: oldTime,
		Type:      monitor.EventForegroundSwitch,
		Package:   "com.ancient.app",
		Details:   map[string]any{},
	})

	// Call MarkUserPopup now
	mgr.MarkUserPopup(now)

	snap := mgr.Snapshot()
	// Ancient app should NOT have been boosted to >= 85 by MarkUserPopup
	for _, c := range snap.Candidates {
		if c.Package == "com.ancient.app" && c.Score >= 85 {
			t.Fatalf("expected com.ancient.app not to be boosted by MarkUserPopup outside 3s window, score: %d", c.Score)
		}
	}

	// 2. Record an event that happened 1.5 seconds ago
	recentTime := now.Add(-1500 * time.Millisecond).Format(time.RFC3339)
	mgr.RecordEvent(monitor.MonitorEvent{
		ID:        "evt-recent",
		Timestamp: recentTime,
		Type:      monitor.EventOverlayShown,
		Package:   "com.recent.popupapp",
		Details:   map[string]any{},
	})

	// Call MarkUserPopup now
	mgr.MarkUserPopup(now)

	snap2 := mgr.Snapshot()
	found := false
	for _, c := range snap2.Candidates {
		if c.Package == "com.recent.popupapp" {
			found = true
			if c.Score < 90 || c.Confidence != "high" {
				t.Fatalf("expected high confidence and score >= 90 for recent overlay event, got score: %d, confidence: %s", c.Score, c.Confidence)
			}
		}
	}
	if !found {
		t.Fatalf("expected candidate com.recent.popupapp in snapshot")
	}
}

func TestMonitorSourcesActiveAccurate(t *testing.T) {
	fake := adb.NewFakeRunner()
	logger := logx.GetDefault()
	mgr := monitor.NewManager(fake, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := mgr.Start(ctx, "dev-1"); err != nil {
		t.Fatalf("failed to start monitor: %v", err)
	}
	defer mgr.Stop()

	status := mgr.Status()
	if !status.IsActive {
		t.Fatalf("expected active monitor")
	}

	// Must be accurately ["dumpsys window"] and not claim unverified ["logcat"]
	if len(status.SourcesActive) != 1 || status.SourcesActive[0] != "dumpsys window" {
		t.Fatalf("expected sourcesActive ['dumpsys window'], got %v", status.SourcesActive)
	}
}

