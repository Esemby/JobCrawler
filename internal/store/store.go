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
}
