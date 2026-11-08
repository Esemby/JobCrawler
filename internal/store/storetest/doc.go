// Package storetest holds the behaviour suite every Store implementation
// must pass, exactly once (docs/testing.md: store tests are written against
// the interface). Engines plug in from their own test files via the open
// seam; no engine knows the suite's internals.
package storetest
