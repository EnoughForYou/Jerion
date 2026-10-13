package audit

import (
	"context"

	"github.com/EnoughForYou/Jerion/internal/adb"
	"github.com/EnoughForYou/Jerion/internal/logx"
)

// AuditCategory categorizes the system security assessment domain.
type AuditCategory string

const (
	CategoryAccessibility        AuditCategory = "accessibility"
	CategoryNotificationListener AuditCategory = "notification_listener"
	CategoryDeviceAdmin          AuditCategory = "device_admin"
	CategoryOverlay              AuditCategory = "overlay"
	CategoryAppOps               AuditCategory = "appops"
	CategoryNetworkSecurity      AuditCategory = "network_security"
	CategoryDevSettings          AuditCategory = "dev_settings"
)

// AuditStatus defines posture verdict.
type AuditStatus string

const (
	StatusOk           AuditStatus = "ok"
	StatusWarning      AuditStatus = "warning"
	StatusCritical     AuditStatus = "critical"
	StatusNotAvailable AuditStatus = "not_available"
)

// AuditItem represents a security posture item matching contract v1.0.0.
type AuditItem struct {
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	Description      string        `json:"description"`
	Category         AuditCategory `json:"category"`
	Status           AuditStatus   `json:"status"`
	CurrentValue     string        `json:"currentValue"`
	RecommendedValue string        `json:"recommendedValue"`
	Remediation      *string       `json:"remediation,omitempty"`
	ActionType       *string       `json:"actionType,omitempty"`
}

// Collector interface for an individual system audit assessment.
type Collector interface {
	Name() string
	Collect(ctx context.Context, runner adb.Runner, serial string) (*AuditItem, error)
}

// Service coordinates system security audits.
type Service struct {
	runner     adb.Runner
	logger     *logx.Logger
	collectors []Collector
}

// NewService creates an audit Service populated with standard collectors.
func NewService(runner adb.Runner, logger *logx.Logger) *Service {
	if logger == nil {
		logger = logx.GetDefault()
	}
	s := &Service{
		runner: runner,
		logger: logger,
	}
	s.registerDefaultCollectors()
	return s
}

func (s *Service) registerDefaultCollectors() {
	s.collectors = []Collector{
		&AccessibilityCollector{},
		&NotificationListenerCollector{},
		&DeviceAdminCollector{},
		&UnknownSourcesCollector{},
		&USBDebuggingCollector{},
		&PrivateDNSCollector{},
		&ProxyCollector{},
	}
}

// RunAudit executes all registered collectors and returns the audit checklist.
func (s *Service) RunAudit(ctx context.Context, serial string) ([]AuditItem, error) {
	var items []AuditItem

	for _, c := range s.collectors {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		item, err := c.Collect(ctx, s.runner, serial)
		if err != nil {
			// Per AGENTS.md §8: command failed / unsupported -> status "not_available", not error
			desc := "Assessment could not be gathered on this device/OEM configuration."
			notAvail := AuditItem{
				ID:               c.Name(),
				Title:            c.Name(),
				Description:      desc,
				Category:         CategoryDevSettings,
				Status:           StatusNotAvailable,
				CurrentValue:     "Unavailable",
				RecommendedValue: "N/A",
			}
			items = append(items, notAvail)
			continue
		}
		if item != nil {
			items = append(items, *item)
		}
	}

	return items, nil
}
