package store

import "context"

// Store is the persistence boundary (docs/architecture.md). Phase 2 scope is
// runtime settings; job/source/verdict methods arrive with later phases and
// will be added to this one interface, never replaced.
//
// Implementations must be safe for use from HTTP handlers: context comes from
// the request, errors are wrapped with context, Close is idempotent.
type Store interface {
	// GetSetting returns the value for key and whether it was found.
	// Not found is not an error: found=false, err=nil.
	GetSetting(ctx context.Context, key string) (value string, found bool, err error)

	// SetSetting upserts key=value (create + update in one call).
	SetSetting(ctx context.Context, key, value string) error

	// ListSettings returns all stored key/value pairs. An empty map means
	// "no runtime overrides"; defaults live in internal/config.
	ListSettings(ctx context.Context) (map[string]string, error)

	// DeleteSetting removes a key. Deleting an absent key succeeds.
	DeleteSetting(ctx context.Context, key string) error

	// Close releases the database handle. It is idempotent and only called
	// at process shutdown.
	Close() error

	// --- sources (Phases 3+) -----------------------------------------------
	// AddSource registers one configured crawler input and returns its id.
	// Duplicate names error. See docs/sources.md for the politeness contract.
	AddSource(ctx context.Context, name, kind, endpoint string) (int64, error)

	// ListSources returns every configured crawler input, sorted by name.
	ListSources(ctx context.Context) ([]Source, error)

	// SetSourceEnabled flips one source by name; unknown names error.
	SetSourceEnabled(ctx context.Context, name string, enabled bool) error

	// --- jobs ---------------------------------------------------------------
	// UpsertJobs stores jobs by fingerprint and returns how many were new.
	UpsertJobs(ctx context.Context, jobs []Job) (int, error)

	// ListJobs returns stored jobs, newest last_seen first. state filters
	// ('active', 'disappeared'); pass "*" for everything.
	ListJobs(ctx context.Context, state string) ([]Job, error)

	// MarkJobState moves one fingerprint's state; 'active'/'disappeared'
	// only; anything else errors.
	MarkJobState(ctx context.Context, fingerprint, state string) error
}
