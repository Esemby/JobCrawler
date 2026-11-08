package storetest

import (
	"context"
	"reflect"
	"testing"

	"github.com/Esemby/JobCrawler/internal/store"
)

// Run executes the settings-CRUD contract suite. open must return a fresh,
// migrated store on an isolated database; the suite closes each instance.
func Run(t *testing.T, open func(t *testing.T) store.Store) {
	t.Helper()
	cases := []struct {
		name string
		run  func(t *testing.T, s store.Store)
	}{
		{"get missing is not found", runGetMissingIsNotFound},
		{"set then get roundtrip", runSetThenGet},
		{"set overwrites existing", runSetOverwrites},
		{"empty store lists nothing", runListEmpty},
		{"list returns all overrides", runListReturnsAll},
		{"delete removes a key", runDeleteRemoves},
		{"delete missing key succeeds", runDeleteMissing},
		{"close twice is safe", runCloseTwice},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := open(t)
			t.Cleanup(func() { _ = s.Close() })
			tc.run(t, s)
		})
	}
}

func runGetMissingIsNotFound(t *testing.T, s store.Store) {
	ctx := context.Background()
	value, found, err := s.GetSetting(ctx, "system.log_level")
	if err != nil {
		t.Fatalf("GetSetting() err = %v, want nil", err)
	}
	if found {
		t.Errorf("GetSetting() found = true, want false for absent key")
	}
	if value != "" {
		t.Errorf("GetSetting() value = %q, want empty for absent key", value)
	}
}

func runSetThenGet(t *testing.T, s store.Store) {
	ctx := context.Background()
	if err := s.SetSetting(ctx, "system.log_level", "debug"); err != nil {
		t.Fatalf("SetSetting() err = %v, want nil", err)
	}
	value, found, err := s.GetSetting(ctx, "system.log_level")
	if err != nil || !found {
		t.Fatalf("GetSetting() = (%q, %v, %v), want found with nil err", value, found, err)
	}
	if value != "debug" {
		t.Errorf("GetSetting() value = %q, want \"debug\"", value)
	}
}

func runSetOverwrites(t *testing.T, s store.Store) {
	ctx := context.Background()
	if err := s.SetSetting(ctx, "system.log_level", "warn"); err != nil {
		t.Fatalf("SetSetting() err = %v, want nil", err)
	}
	if err := s.SetSetting(ctx, "system.log_level", "error"); err != nil {
		t.Fatalf("SetSetting() overwrite err = %v, want nil", err)
	}
	value, found, err := s.GetSetting(ctx, "system.log_level")
	if err != nil || !found {
		t.Fatalf("GetSetting() = (%q, %v, %v), want found with nil err", value, found, err)
	}
	if value != "error" {
		t.Errorf("GetSetting() value = %q, want \"error\" (last write wins)", value)
	}
}

func runListEmpty(t *testing.T, s store.Store) {
	got, err := s.ListSettings(context.Background())
	if err != nil {
		t.Fatalf("ListSettings() err = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Errorf("ListSettings() = %v, want empty map on fresh store", got)
	}
}

func runListReturnsAll(t *testing.T, s store.Store) {
	ctx := context.Background()
	want := map[string]string{
		"system.log_level":  "debug",
		"system.log_format": "json",
		"crawl.concurrency": "4",
	}
	for k, v := range want {
		if err := s.SetSetting(ctx, k, v); err != nil {
			t.Fatalf("SetSetting(%q) err = %v, want nil", k, err)
		}
	}
	got, err := s.ListSettings(ctx)
	if err != nil {
		t.Fatalf("ListSettings() err = %v, want nil", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ListSettings() = %v, want %v", got, want)
	}
}

func runDeleteRemoves(t *testing.T, s store.Store) {
	ctx := context.Background()
	if err := s.SetSetting(ctx, "system.log_level", "debug"); err != nil {
		t.Fatalf("SetSetting() err = %v, want nil", err)
	}
	if err := s.DeleteSetting(ctx, "system.log_level"); err != nil {
		t.Fatalf("DeleteSetting() err = %v, want nil", err)
	}
	_, found, err := s.GetSetting(ctx, "system.log_level")
	if err != nil {
		t.Fatalf("GetSetting() err = %v, want nil", err)
	}
	if found {
		t.Errorf("GetSetting() found = true after Delete, want false")
	}
}

func runDeleteMissing(t *testing.T, s store.Store) {
	if err := s.DeleteSetting(context.Background(), "no.such.key"); err != nil {
		t.Errorf("DeleteSetting() on absent key err = %v, want nil", err)
	}
}

func runCloseTwice(t *testing.T, s store.Store) {
	if err := s.Close(); err != nil {
		t.Fatalf("first Close() err = %v, want nil", err)
	}
	if err := s.Close(); err != nil {
		t.Errorf("second Close() err = %v, want nil (idempotent)", err)
	}
}
