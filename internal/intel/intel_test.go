package intel

import (
	"context"
	"errors"
	"testing"

	"github.com/EnoughForYou/Jerion/internal/secrets"
	"github.com/EnoughForYou/Jerion/internal/store"
)

func TestLocalIntelProvider(t *testing.T) {
	s, err := store.OpenMemory()
	if err != nil {
		t.Fatalf("OpenMemory failed: %v", err)
	}
	defer s.Close()

	hashHex := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	_ = s.UpsertIntelHash(store.IntelHashRecord{
		SHA256:         hashHex,
		Classification: "adware",
		ThreatName:     "Android.Adware.MobiDash",
		Source:         "unit_test",
	})

	provider := NewLocalProvider(s)

	res, err := provider.Lookup(context.Background(), hashHex)
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	if res.Classification != ClassAdware || res.ThreatName != "Android.Adware.MobiDash" {
		t.Errorf("unexpected intel result: %+v", res)
	}

	_, err = provider.Lookup(context.Background(), "unknown_hash_123")
	if !errors.Is(err, ErrHashNotFound) {
		t.Errorf("expected ErrHashNotFound, got: %v", err)
	}
}

func TestIntelAggregator(t *testing.T) {
	s, _ := store.OpenMemory()
	defer s.Close()

	secMgr := secrets.NewManager(secrets.NewMemoryKeyring())
	agg := NewAggregator(s, secMgr)

	_, err := agg.Lookup(context.Background(), "0000000000000000000000000000000000000000000000000000000000000000")
	if !errors.Is(err, ErrHashNotFound) {
		t.Errorf("expected ErrHashNotFound, got %v", err)
	}
}
