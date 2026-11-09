package web

import (
	"github.com/Esemby/JobCrawler/internal/config"
	"github.com/Esemby/JobCrawler/internal/store"
)

// Deps are the Phase 2 runtime dependencies the settings pages need.
// Everything optional behavior hangs off small interfaces so tests can plug
// fakes; see docs/architecture.md (web may reach everything inward).
type Deps struct {
	// Store persists runtime setting overrides. Settings CRUD only.
	Store store.Store
	// Bootstrap is the loaded storage-engine layer (ADR-0002); the Database
	// page displays it and saves a new file via config.WriteBootstrap.
	Bootstrap config.Bootstrap
	// SavedHook runs after every successful settings save so main can apply
	// hot-reloadable values (e.g. log level). It receives the effective
	// merged settings. May be nil in tests.
	SavedHook func(effective map[string]string) error
}
