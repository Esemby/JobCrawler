package greenhouse

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Esemby/JobCrawler/internal/store"
)

// Layout mirrors the official Greenhouse job-board list response
// (developers.greenhouse.io/job-board.html, "List jobs"). Only the fields we
// actually consume are declared; unknown fields are ignored on purpose so a
// company adding custom metadata cannot break parsing.
type Layout struct {
	Jobs []item `json:"jobs"`
	Meta struct {
		Total int `json:"total"`
	} `json:"meta"`
}

type item struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	UpdatedAt    string `json:"updated_at"`
	AbsoluteURL  string `json:"absolute_url"`
	Language     string `json:"language"`
	LocationHint struct {
		Name string `json:"name"`
	} `json:"location"`
}

// Parse validates a job-board list response into a Layout. Unknown fields are
// ignored on purpose so a company adding custom metadata cannot break
// parsing. Empty jobs array is an error, not silence: a board that disables
// its API must not look like an empty crawl.
func Parse(data []byte, boardToken string) (layout Layout, err error) {
	if boardToken == "" {
		return Layout{}, fmt.Errorf("parse greenhouse: boardToken must not be empty")
	}
	if err := json.Unmarshal(data, &layout); err != nil {
		return Layout{}, fmt.Errorf("parse greenhouse %s: %w", boardToken, err)
	}
	if len(layout.Jobs) == 0 {
		return Layout{}, fmt.Errorf("parse greenhouse %s: zero job records", boardToken)
	}
	for i, it := range layout.Jobs {
		if strings.TrimSpace(it.Title) == "" || strings.TrimSpace(it.AbsoluteURL) == "" {
			return Layout{}, fmt.Errorf("parse greenhouse %s: record %d lacks title or absolute_url", boardToken, i)
		}
	}
	return layout, nil
}

// Endpoint is the canonical listing URL for one board.
func Endpoint(boardToken string) string {
	if boardToken == "" || strings.ContainsAny(boardToken, " /?#") {
		return ""
	}
	return "https://boards-api.greenhouse.io/v1/boards/" + boardToken + "/jobs"
}

// ToJobs maps a validated layout into store.Job records for the given source
// id. Company is passed by the caller (a board token does not name the legal
// company, see docs/sources.md); at crawl time it defaults to the token.
func ToJobs(l Layout, sourceID int64, company string) []store.Job {
	out := make([]store.Job, 0, len(l.Jobs))
	for _, it := range l.Jobs {
		out = append(out, store.Job{
			SourceID:    sourceID,
			Title:       strings.TrimSpace(it.Title),
			Company:     company,
			Location:    strings.TrimSpace(it.LocationHint.Name),
			URL:         strings.TrimSpace(it.AbsoluteURL),
			PublishedAt: it.UpdatedAt,
		})
	}
	return out
}
