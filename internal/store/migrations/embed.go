// Package migrations holds the numbered SQL migrations, embedded in the
// binary. Files follow docs/playbooks/add-a-migration.md: description-only
// names, plain SQL, never edited after being applied.
//
// Embed only — no logic. The runner lives in the store package so both
// engines reuse it.
package migrations

import "embed"

// FS holds every migration file: NNNN_description.sql (shared dialect) and
// NNNN_description.<engine>.sql when engines must differ. The runner picks
// by suffix, falling back to the engine-less file.
//
//go:embed *.sql
var FS embed.FS
