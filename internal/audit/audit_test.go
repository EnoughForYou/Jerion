package audit

import (
	"context"
	"testing"

	"github.com/EnoughForYou/Jerion/internal/adb"
)

func TestAuditService(t *testing.T) {
	fake := adb.NewFakeRunner()

	// 1. Accessibility: enabled services present
	fake.SetResponse("settings get secure enabled_accessibility_services", "com.cleaner.master/com.cleaner.master.AccService\n", "", nil)

	// 2. Notification listeners: none
	fake.SetResponse("cmd notification get_listeners", "No listeners found\n", "", nil)

	// 3. Device Admin: active admin
	fake.SetResponse("dpm list-owners", "admin=com.suspicious.lock/com.suspicious.lock.AdminReceiver\n", "", nil)

	// 4. Unknown sources: disabled
	fake.SetResponse("settings get secure install_non_market_apps", "0\n", "", nil)

	// 5. USB debugging: enabled
	fake.SetResponse("settings get global adb_enabled", "1\n", "", nil)

	// 6. Private DNS: strict
	fake.SetResponse("settings get global private_dns_mode", "hostname\n", "", nil)

	// 7. Proxy: none
	fake.SetResponse("settings get global http_proxy", ":0\n", "", nil)

	svc := NewService(fake, nil)
	items, err := svc.RunAudit(context.Background(), "DEV-01")
	if err != nil {
		t.Fatalf("RunAudit failed: %v", err)
	}

	if len(items) != 7 {
		t.Fatalf("expected 7 audit items, got %d", len(items))
	}

	itemMap := make(map[string]AuditItem)
	for _, it := range items {
		itemMap[it.ID] = it
	}

	// Accessibility should be warning
	acc := itemMap["audit-accessibility"]
	if acc.Status != StatusWarning {
		t.Errorf("expected accessibility StatusWarning, got %s", acc.Status)
	}

	// Notification listeners should be ok
	notif := itemMap["audit-notification-listeners"]
	if notif.Status != StatusOk {
		t.Errorf("expected notification listener StatusOk, got %s", notif.Status)
	}

	// Device admin should be critical
	admin := itemMap["audit-device-admin"]
	if admin.Status != StatusCritical {
		t.Errorf("expected device admin StatusCritical, got %s", admin.Status)
	}

	// Unknown sources should be ok
	unk := itemMap["audit-unknown-sources"]
	if unk.Status != StatusOk {
		t.Errorf("expected unknown sources StatusOk, got %s", unk.Status)
	}
}
