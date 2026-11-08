// Package version holds build-time version information for the binary.
//
// It is set at build time via the Makefile ldflags; unknown values mean the
// binary was built outside `make build`. Keep this package free of imports.
package version
