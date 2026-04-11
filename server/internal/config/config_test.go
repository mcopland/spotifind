package config

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureSlog redirects the default slog logger to a buffer for the duration
// of the test and returns the buffer. The previous default logger is restored
// on cleanup. Tests that call this must not run in parallel.
func captureSlog(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return buf
}

// chdir switches CWD to dir for the duration of the test and restores the
// original on cleanup. Tests that call this must not run in parallel.
func chdir(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %q: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(prev); err != nil {
			t.Fatalf("restore cwd %q: %v", prev, err)
		}
	})
}

// TestLoad_MissingDotEnvFile asserts that Load succeeds when no .env file
// exists at the expected ../.env path, as long as the required environment
// variables are set. This guards the "missing .env is not fatal" contract
// that the godotenv error handling depends on.
func TestLoad_MissingDotEnvFile(t *testing.T) {
	root := t.TempDir()
	workdir := filepath.Join(root, "server")
	if err := os.Mkdir(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}
	chdir(t, workdir)

	t.Setenv("DATABASE_URL", "postgres://localhost/db")
	t.Setenv("SPOTIFY_CLIENT_ID", "cid")
	t.Setenv("SPOTIFY_CLIENT_SECRET", "csecret")
	t.Setenv("JWT_SECRET", "jwt")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DatabaseURL != "postgres://localhost/db" {
		t.Errorf("DatabaseURL: want postgres://localhost/db, got %q", cfg.DatabaseURL)
	}
}

// TestLoad_DotEnvLoadErrorIsLoggedNotSwallowed asserts that when ../.env
// exists but cannot be parsed as a dotenv file (here, because it is a
// directory rather than a regular file), Load tolerates the failure but
// surfaces it via a warning log. The previous behavior discarded the error
// with `_ =`, hiding misconfigured `.env` files entirely. Load itself must
// still succeed because the required env vars are present in the process
// environment.
func TestLoad_DotEnvLoadErrorIsLoggedNotSwallowed(t *testing.T) {
	root := t.TempDir()
	workdir := filepath.Join(root, "server")
	if err := os.Mkdir(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, ".env"), 0o755); err != nil {
		t.Fatalf("mkdir .env: %v", err)
	}
	chdir(t, workdir)

	logs := captureSlog(t)

	t.Setenv("DATABASE_URL", "postgres://localhost/db")
	t.Setenv("SPOTIFY_CLIENT_ID", "cid")
	t.Setenv("SPOTIFY_CLIENT_SECRET", "csecret")
	t.Setenv("JWT_SECRET", "jwt")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DatabaseURL != "postgres://localhost/db" {
		t.Errorf("DatabaseURL: want postgres://localhost/db, got %q", cfg.DatabaseURL)
	}

	out := logs.String()
	if !strings.Contains(out, "level=WARN") {
		t.Errorf("expected a WARN log when godotenv fails, got: %q", out)
	}
	if !strings.Contains(out, "failed to load .env") {
		t.Errorf("expected log to name the failure, got: %q", out)
	}
}
