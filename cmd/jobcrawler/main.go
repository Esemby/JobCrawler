// Command jobcrawler wires configuration, version info, the store and the web
// server.
//
// main.go is wiring only: no business logic, per docs/architecture.md. Later
// phases add scheduler/scan wiring here in the same thin style.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Esemby/JobCrawler/internal/config"
	"github.com/Esemby/JobCrawler/internal/store/sqlite"
	"github.com/Esemby/JobCrawler/internal/version"
	"github.com/Esemby/JobCrawler/internal/web"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	slog.SetDefault(newLogger(cfg))
	path := os.Getenv("JOBCRAWLER_CONFIG")
	if path == "" {
		path = config.DefaultBootstrapPath
	}
	boot, err := config.LoadBootstrap(path)
	if err != nil {
		return fmt.Errorf("bootstrap %s: %w", path, err)
	}
	if boot.Engine != "sqlite" {
		return fmt.Errorf("engine %q: mariadb store arrives in Phase 6; set engine to sqlite", boot.Engine)
	}

	ctx := context.Background()
	st, err := sqlite.Open(ctx, boot.DSN)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer func() {
		if err := st.Close(); err != nil {
			slog.Error("store close", "err", err)
		}
	}()

	// Runtime overrides stored in the DB win over defaults (docs/config.md):
	// apply the merged values before building pages so the hot hook matches.
	overrides, _ := st.ListSettings(ctx)
	effective := config.Effective(overrides)
	slog.SetDefault(loggerFor(effective))

	srv, err := web.New(web.StatusPage{
		Version:   version.Version,
		Commit:    version.Commit,
		Date:      version.Date,
		GoVersion: runtime.Version(),
		BindAddr:  cfg.BindAddr,
		DBEngine:  boot.Engine,
	}, web.Deps{Store: st, Bootstrap: boot, SavedHook: hookFor(effective)})
	if err != nil {
		return fmt.Errorf("web: %w", err)
	}

	httpSrv := &http.Server{
		Addr:              cfg.BindAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpSrv.ListenAndServe()
	}()
	slog.Info("ui listening", "addr", cfg.BindAddr, "db", fmt.Sprintf("sqlite %s", boot.DSN))

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve: %w", err)
		}
	case <-sigCtx.Done():
		slog.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			slog.Error("http shutdown", "err", err)
		}
	}
	return nil
}

// loggerFor builds the default logger from the base config adjusted by any
// effective (DB-stored) level override; format stays startup-fixed.
func loggerFor(effective map[string]string) *slog.Logger {
	level, err := config.ParseLogLevel(effective["system.log_level"])
	if err != nil {
		level = slog.LevelInfo
	}
	return buildLogger(level, os.Getenv("JOBCRAWLER_LOG_FORMAT"))
}

// hookFor returns the SavedHook built over the startup config so every saved
// setting re-applies "live" values without restart.
func hookFor(base map[string]string) func(map[string]string) error {
	return func(effective map[string]string) error {
		slog.SetDefault(loggerFor(effective))
		_ = base
		return nil
	}
}

func newLogger(cfg config.Config) *slog.Logger {
	level, err := config.ParseLogLevel(cfg.LogLevel)
	if err != nil { // Load validated already; belt and braces for direct callers
		level = slog.LevelInfo
	}
	return buildLogger(level, cfg.LogFormat)
}

func buildLogger(level slog.Level, format string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: level}
	if format == "json" {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}
