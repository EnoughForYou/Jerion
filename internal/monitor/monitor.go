package monitor

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/EnoughForYou/Jerion/internal/adb"
	"github.com/EnoughForYou/Jerion/internal/logx"
)

// EventType enumerates monitor signals matching contract v1.0.0.
type EventType string

const (
	EventForegroundSwitch   EventType = "foreground_switch"
	EventOverlayShown       EventType = "overlay_shown"
	EventNotificationPosted EventType = "notification_posted"

	// Backward-compatibility aliases
	EventWindowFocusChange  EventType = EventForegroundSwitch
	EventPackageStateChange EventType = EventForegroundSwitch
	EventLogcatTrigger      EventType = EventNotificationPosted
)

// MonitorStatus represents real-time daemon state.
type MonitorStatus struct {
	IsActive      bool     `json:"isActive"`
	StartedAt     *string  `json:"startedAt,omitempty"`
	SourcesActive []string `json:"sourcesActive"`
}

// MonitorEvent represents a telemetry signal.
type MonitorEvent struct {
	ID        string         `json:"id"`
	Timestamp string         `json:"timestamp"`
	Type      EventType      `json:"type"`
	Package   string         `json:"package"`
	Details   map[string]any `json:"details"`
}

// MonitorCandidate represents an identified adware / popup suspect.
type MonitorCandidate struct {
	Package        string         `json:"package"`
	AppLabel       string         `json:"appLabel"`
	Confidence     string         `json:"confidence"` // "low" | "medium" | "high"
	Score          int            `json:"score"`
	TimelineEvents []MonitorEvent `json:"timelineEvents"`
	WhyFlagged     string         `json:"whyFlagged"`
}

// MonitorSnapshot captures instantaneous daemon telemetry.
type MonitorSnapshot struct {
	Status     MonitorStatus      `json:"status"`
	Events     []MonitorEvent     `json:"events"`
	Candidates []MonitorCandidate `json:"candidates"`
}

// EventCallback receives real-time stream events.
type EventCallback func(evt MonitorEvent)

// CandidateCallback receives newly flagged or re-scored candidates.
type CandidateCallback func(cand MonitorCandidate)

// Manager coordinates real-time monitoring and popup attribution.
type Manager struct {
	runner        adb.Runner
	logger        *logx.Logger
	mu            sync.RWMutex
	isActive      bool
	startedAt     *time.Time
	activeSerial  string
	events        []MonitorEvent
	candidates    map[string]*MonitorCandidate
	eventCb       []EventCallback
	candidateCb   []CandidateCallback
	cancelMonitor context.CancelFunc
}

// NewManager creates a popup hunter Manager.
func NewManager(runner adb.Runner, logger *logx.Logger) *Manager {
	if logger == nil {
		logger = logx.GetDefault()
	}
	return &Manager{
		runner:     runner,
		logger:     logger,
		candidates: make(map[string]*MonitorCandidate),
	}
}

// OnEvent registers an event stream listener.
func (m *Manager) OnEvent(cb EventCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventCb = append(m.eventCb, cb)
}

// OnCandidate registers a candidate attribution listener.
func (m *Manager) OnCandidate(cb CandidateCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.candidateCb = append(m.candidateCb, cb)
}

func (m *Manager) statusLocked() MonitorStatus {
	var startedAtStr *string
	if m.startedAt != nil {
		s := m.startedAt.Format(time.RFC3339)
		startedAtStr = &s
	}

	sources := []string{}
	if m.isActive {
		sources = []string{"dumpsys window"}
	}

	return MonitorStatus{
		IsActive:      m.isActive,
		StartedAt:     startedAtStr,
		SourcesActive: sources,
	}
}

// Status returns current daemon activity state.
func (m *Manager) Status() MonitorStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.statusLocked()
}

// Snapshot returns the current collection of events and ranked candidates.
func (m *Manager) Snapshot() MonitorSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := m.statusLocked()

	eventsCopy := make([]MonitorEvent, len(m.events))
	copy(eventsCopy, m.events)

	var candidatesList []MonitorCandidate
	for _, c := range m.candidates {
		candidatesList = append(candidatesList, *c)
	}

	// Sort candidates by score descending
	sort.Slice(candidatesList, func(i, j int) bool {
		return candidatesList[i].Score > candidatesList[j].Score
	})

	return MonitorSnapshot{
		Status:     status,
		Events:     eventsCopy,
		Candidates: candidatesList,
	}
}

// Start begins background telemetry polling.
func (m *Manager) Start(ctx context.Context, serial string) error {
	m.mu.Lock()
	if m.isActive {
		m.mu.Unlock()
		return nil
	}

	now := time.Now().UTC()
	m.isActive = true
	m.startedAt = &now
	m.activeSerial = serial

	monCtx, cancel := context.WithCancel(ctx)
	m.cancelMonitor = cancel
	m.mu.Unlock()

	go m.runCollectorLoop(monCtx, serial)

	return nil
}

// Stop terminates the active monitoring daemon.
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancelMonitor != nil {
		m.cancelMonitor()
		m.cancelMonitor = nil
	}
	m.isActive = false
	m.startedAt = nil
}

// RecordEvent appends an event, trims history to recent 100, and triggers attribution.
func (m *Manager) RecordEvent(evt MonitorEvent) {
	m.mu.Lock()
	m.events = append([]MonitorEvent{evt}, m.events...)
	if len(m.events) > 100 {
		m.events = m.events[:100]
	}
	listeners := m.eventCb
	m.mu.Unlock()

	for _, l := range listeners {
		l(evt)
	}

	// Evaluate candidate attribution
	m.evaluateAttribution(evt)
}

// MarkUserPopup records a manual trigger and correlates recent events (< 3000ms).
func (m *Manager) MarkUserPopup(t time.Time) {
	if t.IsZero() {
		t = time.Now().UTC()
	}

	m.mu.Lock()
	var recentCandidate *MonitorCandidate

	// Find most recent non-system overlay or foreground event within 3 seconds of t
	for _, e := range m.events {
		if e.Package == "" || stringsHasSystemPrefix(e.Package) {
			continue
		}

		evtTime, err := time.Parse(time.RFC3339, e.Timestamp)
		if err != nil {
			continue
		}

		diff := t.Sub(evtTime)
		if diff < 0 {
			diff = -diff
		}
		if diff > 3*time.Second {
			continue
		}

		cand, ok := m.candidates[e.Package]
		if !ok {
			score := 85
			confidence := "medium"
			why := "Foreground activity captured within 3s of manual 'Popup Just Appeared!' trigger."
			if e.Type == EventOverlayShown {
				score = 95
				confidence = "high"
				why = "Active TYPE_APPLICATION_OVERLAY captured within 3s of manual 'Popup Just Appeared!' trigger."
			}

			cand = &MonitorCandidate{
				Package:        e.Package,
				AppLabel:       prettifyPackage(e.Package),
				Confidence:     confidence,
				Score:          score,
				TimelineEvents: []MonitorEvent{e},
				WhyFlagged:     why,
			}
			m.candidates[e.Package] = cand
		} else {
			if e.Type == EventOverlayShown {
				cand.Score = 95
				cand.Confidence = "high"
				cand.WhyFlagged = "Emphasized by user mark trigger during active window overlay (< 3s)."
			} else if cand.Score < 85 {
				cand.Score = 85
				cand.Confidence = "medium"
				cand.WhyFlagged = "Correlated with foreground transition adjacent to user mark trigger (< 3s)."
			}
		}
		recentCandidate = cand
		break
	}
	candListeners := m.candidateCb
	m.mu.Unlock()

	if recentCandidate != nil {
		for _, l := range candListeners {
			l(*recentCandidate)
		}
	}
}

func (m *Manager) evaluateAttribution(evt MonitorEvent) {
	if evt.Package == "" || stringsHasSystemPrefix(evt.Package) {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	cand, ok := m.candidates[evt.Package]
	if !ok {
		score := 40
		confidence := "low"
		why := "Observed frequent foreground focus transitions."

		if evt.Type == EventOverlayShown {
			score = 85
			confidence = "high"
			why = "Displayed TYPE_APPLICATION_OVERLAY window over current activity."
		}

		cand = &MonitorCandidate{
			Package:        evt.Package,
			AppLabel:       prettifyPackage(evt.Package),
			Confidence:     confidence,
			Score:          score,
			TimelineEvents: []MonitorEvent{evt},
			WhyFlagged:     why,
		}
		m.candidates[evt.Package] = cand
	} else {
		cand.TimelineEvents = append([]MonitorEvent{evt}, cand.TimelineEvents...)
		if len(cand.TimelineEvents) > 50 {
			cand.TimelineEvents = cand.TimelineEvents[:50]
		}
		if evt.Type == EventOverlayShown && cand.Score < 85 {
			cand.Score = 85
			cand.Confidence = "high"
			cand.WhyFlagged = "Observed displaying fullscreen TYPE_APPLICATION_OVERLAY window."
		}
	}

	candCopy := *cand
	for _, l := range m.candidateCb {
		l(candCopy)
	}
}

func stringsHasSystemPrefix(pkg string) bool {
	return pkg == "com.android.systemui" || pkg == "com.android.settings" || pkg == "android"
}

func prettifyPackage(pkg string) string {
	parts := splitPackage(pkg)
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		if len(last) > 0 {
			return fmt.Sprintf("%s%s", string(last[0]-32), last[1:])
		}
	}
	return pkg
}

func splitPackage(pkg string) []string {
	var res []string
	start := 0
	for i := 0; i < len(pkg); i++ {
		if pkg[i] == '.' {
			res = append(res, pkg[start:i])
			start = i + 1
		}
	}
	if start < len(pkg) {
		res = append(res, pkg[start:])
	}
	return res
}
