package config

import (
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	config, err := load(func(key string) string {
		return map[string]string{
			databaseURLEnv: "postgres://localhost:5432/tournaments",
			httpAddrEnv:    "127.0.0.1:8080",
		}[key]
	})
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	if config.HTTPAddr != "127.0.0.1:8080" {
		t.Errorf("HTTPAddr = %q, want %q", config.HTTPAddr, "127.0.0.1:8080")
	}
}

func TestLoadRejectsMissingDatabaseURL(t *testing.T) {
	t.Parallel()

	_, err := load(func(key string) string {
		if key == httpAddrEnv {
			return "127.0.0.1:8080"
		}
		return ""
	})
	if err == nil || !strings.Contains(err.Error(), databaseURLEnv) {
		t.Fatalf("load() error = %v, want error mentioning %s", err, databaseURLEnv)
	}
}

func TestLoadRejectsNonPostgreSQLURL(t *testing.T) {
	t.Parallel()

	_, err := load(func(key string) string {
		return map[string]string{
			databaseURLEnv: "https://example.com",
			httpAddrEnv:    "127.0.0.1:8080",
		}[key]
	})
	if err == nil || !strings.Contains(err.Error(), databaseURLEnv) {
		t.Fatalf("load() error = %v, want error mentioning %s", err, databaseURLEnv)
	}
}
