package main

import (
	"testing"
)

// NOTE: This placeholder test verifies the base Go test harness for Phase 0.
// Temporary: this file will be updated/replaced in Phase 4 / Phase 2 as domain services are added.
func TestAppPlaceholder(t *testing.T) {
	app := NewApp()
	if app == nil {
		t.Fatal("expected non-nil App instance")
	}

	expected := "Hello Jerion, It's show time!"
	actual := app.Greet("Jerion")
	if actual != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}
