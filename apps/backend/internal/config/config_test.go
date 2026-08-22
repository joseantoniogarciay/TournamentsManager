package config

import (
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	config, err := load(func(key string) string {
		return map[string]string{
			databaseURLEnv:        "postgres://localhost:5432/tournaments",
			httpAddrEnv:           "127.0.0.1:8080",
			smtpAddrEnv:           "127.0.0.1:1025",
			smtpFromEnv:           "no-reply@example.test",
			publicBaseURLEnv:      "http://127.0.0.1:8080",
			corsAllowedOriginsEnv: "http://localhost:8082, http://127.0.0.1:8082",
		}[key]
	})
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	if config.HTTPAddr != "127.0.0.1:8080" {
		t.Errorf("HTTPAddr = %q, want %q", config.HTTPAddr, "127.0.0.1:8080")
	}
	if len(config.CORSAllowedOrigins) != 2 {
		t.Errorf("CORSAllowedOrigins = %v, want two origins", config.CORSAllowedOrigins)
	}
}

func TestLoadRejectsInvalidCORSOrigin(t *testing.T) {
	t.Parallel()

	_, err := load(func(key string) string {
		return map[string]string{
			databaseURLEnv:        "postgres://localhost:5432/tournaments",
			httpAddrEnv:           "127.0.0.1:8080",
			smtpAddrEnv:           "127.0.0.1:1025",
			smtpFromEnv:           "no-reply@example.test",
			publicBaseURLEnv:      "http://127.0.0.1:8080",
			corsAllowedOriginsEnv: "https://example.test/path",
		}[key]
	})
	if err == nil || !strings.Contains(err.Error(), corsAllowedOriginsEnv) {
		t.Fatalf("load() error = %v, want error mentioning %s", err, corsAllowedOriginsEnv)
	}
}

func TestLoadRejectsNonHTTPSPublicBaseURLOutsideLoopback(t *testing.T) {
	t.Parallel()

	_, err := load(func(key string) string {
		return map[string]string{
			databaseURLEnv:        "postgres://localhost:5432/tournaments",
			httpAddrEnv:           "127.0.0.1:8080",
			smtpAddrEnv:           "127.0.0.1:1025",
			smtpFromEnv:           "no-reply@example.test",
			publicBaseURLEnv:      "http://links.example.test",
			corsAllowedOriginsEnv: "http://localhost:8082",
		}[key]
	})
	if err == nil || !strings.Contains(err.Error(), publicBaseURLEnv) {
		t.Fatalf("load() error = %v, want error mentioning %s", err, publicBaseURLEnv)
	}
}

func TestLoadRejectsIncompleteSMTPCredentials(t *testing.T) {
	t.Parallel()

	_, err := load(func(key string) string {
		return map[string]string{
			databaseURLEnv:        "postgres://localhost:5432/tournaments",
			httpAddrEnv:           "127.0.0.1:8080",
			smtpAddrEnv:           "smtp.example.test:587",
			smtpFromEnv:           "no-reply@example.test",
			smtpUsernameEnv:       "resend",
			publicBaseURLEnv:      "https://example.test",
			corsAllowedOriginsEnv: "https://example.test",
		}[key]
	})
	if err == nil || !strings.Contains(err.Error(), smtpPasswordEnv) {
		t.Fatalf("load() error = %v, want error mentioning %s", err, smtpPasswordEnv)
	}
}

func TestLoadRejectsEmailSubjectPrefixWithNewline(t *testing.T) {
	t.Parallel()

	_, err := load(func(key string) string {
		return map[string]string{
			databaseURLEnv:        "postgres://localhost:5432/tournaments",
			httpAddrEnv:           "127.0.0.1:8080",
			smtpAddrEnv:           "127.0.0.1:1025",
			smtpFromEnv:           "no-reply@example.test",
			publicBaseURLEnv:      "http://127.0.0.1:8080",
			corsAllowedOriginsEnv: "http://localhost:8082",
			emailSubjectPrefixEnv: "[DEV]\r\ninjected",
		}[key]
	})
	if err == nil || !strings.Contains(err.Error(), emailSubjectPrefixEnv) {
		t.Fatalf("load() error = %v, want error mentioning %s", err, emailSubjectPrefixEnv)
	}
}

func TestLoadRejectsInvalidOTELTracesEndpoint(t *testing.T) {
	t.Parallel()

	_, err := load(func(key string) string {
		return map[string]string{
			databaseURLEnv:        "postgres://localhost:5432/tournaments",
			httpAddrEnv:           "127.0.0.1:8080",
			smtpAddrEnv:           "127.0.0.1:1025",
			smtpFromEnv:           "no-reply@example.test",
			publicBaseURLEnv:      "http://127.0.0.1:8080",
			corsAllowedOriginsEnv: "http://localhost:8082",
			otelTracesEndpointEnv: "ftp://tempo:4318",
		}[key]
	})
	if err == nil || !strings.Contains(err.Error(), otelTracesEndpointEnv) {
		t.Fatalf("load() error = %v, want error mentioning %s", err, otelTracesEndpointEnv)
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
