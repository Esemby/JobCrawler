package config

import (
	"reflect"
	"strings"
	"testing"
)

// writeBootstrap creates a temp bootstrap file with the given body and
// returns both a default-path copy and an explicit-path copy, to exercise
// both missing-file behaviours.
func writeBootstrap(t *testing.T, body string) string {
	t.Helper()
	path := t.TempDir() + "/config.bootstrap.toml"
	if err := writeFile(path, body); err != nil {
		t.Fatalf("write bootstrap fixture: %v", err)
	}
	return path
}

func TestLoadBootstrapShapes(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    Bootstrap
		wantErr string
	}{
		{
			name: "valid sqlite file",
			body: "# bootstrap (ADR-0002)\n[db]\nengine = \"sqlite\"\ndsn = \"jc.db\"\n",
			want: Bootstrap{Engine: "sqlite", DSN: "jc.db"},
		},
		{
			name: "blank lines and comments tolerated",
			body: "\n# nothing yet\n\n[db]\n# engine omitted, defaults apply\ndsn = \"x.db\"\n\n",
			want: Bootstrap{Engine: "sqlite", DSN: "x.db"},
		},
		{
			name: "mariadb with dsn",
			body: "[db]\nengine = \"mariadb\"\ndsn = \"jobcrawler:not-on-git@127.0.0.1:3306/jc\"\n",
			want: Bootstrap{Engine: "mariadb", DSN: "jobcrawler:not-on-git@127.0.0.1:3306/jc"},
		},
		{
			name:    "key outside [db] table",
			body:    "engine = \"sqlite\"\n",
			wantErr: "must appear under the [db] table",
		},
		{
			name:    "unknown table",
			body:    "[crawl]\n",
			wantErr: `unknown table [crawl]`,
		},
		{
			name:    "unknown key",
			body:    "[db]\nengin = \"sqlite\"\n",
			wantErr: `unknown key "engin"`,
		},
		{
			name:    "unquoted value",
			body:    "[db]\ndsn = jobcrawler.db\n",
			wantErr: `value for "dsn" must be a double-quoted string`,
		},
		{
			name:    "no equals sign",
			body:    "[db]\ndsn \"jobcrawler.db\"\n",
			wantErr: "missing key = \"value\" pair",
		},
		{
			name:    "bad engine",
			body:    "[db]\nengine = \"postgres\"\ndsn = \"x\"\n",
			wantErr: "db.engine \"postgres\"",
		},
		{
			name:    "mariadb without dsn",
			body:    "[db]\nengine = \"mariadb\"\n",
			wantErr: "non-empty DSN required when engine is mariadb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeBootstrap(t, tt.body)
			got, err := LoadBootstrap(path)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("LoadBootstrap() err = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadBootstrap() err = %v, want nil", err)
			}
			tt.want.Path = path
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadBootstrap() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestLoadBootstrapErrorsWhenExplicitFileMissing(t *testing.T) {
	_, err := LoadBootstrap(t.TempDir() + "/absent.toml")
	if err == nil {
		t.Fatalf("LoadBootstrap(absent) err = nil, want error")
	}
	if !strings.Contains(err.Error(), "absent.toml") {
		t.Errorf("error %v does not name the path", err)
	}
}

func TestLoadBootstrapDefaultPathMissingUsesDefaults(t *testing.T) {
	// A fresh install has no config.bootstrap.toml: documented defaults apply.
	// DefaultBootstrapPath is relative to the process cwd, so the test moves
	// into a private empty directory first.
	dir := t.TempDir()
	t.Setenv("JOBCRAWLER_DB_ENGINE", "")
	t.Setenv("JOBCRAWLER_DB_DSN", "")
	restore := chdir(t, dir)
	defer restore()

	cfg, err := LoadBootstrap(DefaultBootstrapPath)
	if err != nil {
		t.Fatalf("LoadBootstrap(default, missing file) err = %v, want defaults", err)
	}
	if cfg.Engine != "sqlite" || cfg.DSN != "jobcrawler.db" {
		t.Errorf("fresh-install defaults = %+v, want sqlite + jobcrawler.db", cfg)
	}
}

func TestLoadBootstrapEnvOverrides(t *testing.T) {
	t.Setenv("JOBCRAWLER_DB_ENGINE", "mariadb")
	t.Setenv("JOBCRAWLER_DB_DSN", "env-dsn@127.0.0.1/jc")
	path := writeBootstrap(t, "[db]\nengine = \"sqlite\"\ndsn = \"file.db\"\n")

	got, err := LoadBootstrap(path)
	if err != nil {
		t.Fatalf("LoadBootstrap() err = %v, want nil (env beats file)", err)
	}
	if got.Engine != "mariadb" || got.DSN != "env-dsn@127.0.0.1/jc" {
		t.Errorf("env overrides not applied: %+v", got)
	}
}

func TestLoadBootstrapEnvOverridesInvalidEngine(t *testing.T) {
	t.Setenv("JOBCRAWLER_DB_ENGINE", "postgres")
	path := writeBootstrap(t, "[db]\ndsn = \"x.db\"\n")
	_, err := LoadBootstrap(path)
	if err == nil || !strings.Contains(err.Error(), "db.engine \"postgres\"") {
		t.Errorf("LoadBootstrap() err = %v, want db.engine error", err)
	}
}
