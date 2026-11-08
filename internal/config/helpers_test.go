package config

import (
	"fmt"
	"os"
	"testing"
)

// writeFile is a test helper creating a fixture file.
func writeFile(path, body string) error {
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// chdir moves the process into dir for the duration of the test (needed when
// a function resolves a relative default path from the working directory).
// The returned func restores the original directory.
func chdir(t *testing.T, dir string) func() {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	return func() {
		if err := os.Chdir(old); err != nil {
			t.Fatalf("chdir back to %s: %v", old, err)
		}
	}
}
