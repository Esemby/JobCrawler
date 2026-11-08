// Command jobcrawler wires configuration, version info and the web server.
//
// main.go is wiring only: no business logic, per docs/architecture.md. Later
// phases add store/scheduler/scan wiring here in the same thin style.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Esemby/JobCrawler/internal/config"
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
	bootstrapPath := os.Getenv("JOBCRAWLER_CONFIG")
	if bootstrapPath == "" {
		bootstrapPath = config.DefaultBootstrapPath
	}
	bootstrap, err := config.LoadBootstrap(bootstrapPath)
	if err != nil {
		return fmt.Errorf("bootstrap config: %w", err)
	}

	logger := newLogger(cfg)
	slog.SetDefault(logger)
	logger.Info("starting jobcrawler", "version", version.String(), "bind", cfg.BindAddr, "engine", bootstrap.Engine)

	srv, err := web.New(web.StatusPage{
		Version:   version.Version,
		Commit:    version.Commit,
		Date:      version.Date,
		GoVersion: runtime.Version(),
		BindAddr:  cfg.BindAddr,
		DBEngine:  bootstrap.Engine,
	})
	if err != nil {
		return fmt.Errorf("web: %w", err)
	}

	httpSrv := &http.Server{
		Addr:              cfg.BindAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpSrv.ListenAndServe()
	}()
	logger.Info("ui listening", "addr", cfg.BindAddr)

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("serve: %w", err)
		}
	case <-ctx.Done():
		logger.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		logger.Info("shutdown complete", "err", httpSrv.Shutdown(shutdownCtx))
	}
	return nil
}

func newLogger(cfg config.Config) *slog.Logger {
	level := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	opts := &slog.HandlerOptions{Level: level}
	if cfg.LogFormat == "json" {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}
