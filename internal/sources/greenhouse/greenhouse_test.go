package greenhouse

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Esemby/JobCrawler/internal/store"
)

// fixture reads a real saved response (docs/testing.md fixture rules).
func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "greenhouse", name))
	if err != nil {
		t.Fatalf("fixture %s: %v", name, err)
	}
	return b
}

func TestParseBoardFixture(t *testing.T) {
	layout, err := Parse(fixture(t, "board-propublica.json"), "propublica")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if layout.Meta.Total != 4 || len(layout.Jobs) != 4 {
		t.Fatalf("meta.total = %d records = %d, want 4 records from the fixture",
			layout.Meta.Total, len(layout.Jobs))
	}
	first, last := layout.Jobs[0], layout.Jobs[len(layout.Jobs)-1]
	if first.ID == 0 || first.Title == "" || first.AbsoluteURL == "" || first.LocationHint.Name == "" {
		t.Errorf("first record incomplete: %+v", first)
	}
	if last.ID == 0 || last.Title == "" || last.AbsoluteURL == "" {
		t.Errorf("last record incomplete: %+v", last)
	}
}

func TestParseRejectsJunk(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{"not json", "hello", "invalid character"},
		{"empty jobs", `{"jobs":[],"meta":{"total":0}}`, "zero job records"},
		{"missing fields", `{"jobs":[{"id":1}]}`, "lacks title or absolute_url"},
		{"missing token", "", "boardToken"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.data), "acme")
			if tt.name == "missing token" {
				_, err = Parse([]byte(`{"jobs":[{"id":1,"title":"x","absolute_url":"u"}]}`), "")
			}
			if err == nil || !contains(err.Error(), tt.want) {
				t.Errorf("Parse err = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})())
}

func TestToJobs(t *testing.T) {
	layout, err := Parse(fixture(t, "board-propublica.json"), "propublica")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	jobs := ToJobs(layout, 7, "ProPublica")
	if len(jobs) != 4 {
		t.Fatalf("ToJobs length = %d, want 4", len(jobs))
	}
	j := jobs[0]
	if j.SourceID != 7 || j.Company != "ProPublica" ||
		j.Title == "" || j.URL == "" || j.Location == "" {
		t.Errorf("first mapped job incomplete: %+v", j)
	}
	fp := store.Fingerprint(j)
	if !strings.Contains(fp, "propublica") {
		t.Errorf("fingerprint %q missing normalized company", fp)
	}
}

func TestEndpoint(t *testing.T) {
	tests := []struct {
		token, want string
	}{
		{"acme", "https://boards-api.greenhouse.io/v1/boards/acme/jobs"},
		{"", ""},
		{"bad token/slash", ""},
	}
	for _, tt := range tests {
		if got := Endpoint(tt.token); got != tt.want {
			t.Errorf("Endpoint(%q) = %q, want %q", tt.token, got, tt.want)
		}
	}
}
