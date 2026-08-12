// Package config carga y valida la configuración de ejecución de la API.
package config

import (
	"fmt"
	"net/netip"
	"net/url"
	"os"
	"strings"
)

const (
	databaseURLEnv        = "DATABASE_URL"
	httpAddrEnv           = "HTTP_ADDR"
	smtpAddrEnv           = "SMTP_ADDR"
	smtpFromEnv           = "SMTP_FROM"
	publicBaseURLEnv      = "PUBLIC_BASE_URL"
	corsAllowedOriginsEnv = "CORS_ALLOWED_ORIGINS"
	googleClientIDsEnv    = "GOOGLE_CLIENT_IDS"
	trustedProxyCIDRsEnv  = "TRUSTED_PROXY_CIDRS"
)

// Config contiene únicamente la configuración necesaria para arrancar la API.
type Config struct {
	DatabaseURL        string
	HTTPAddr           string
	SMTPAddr           string
	SMTPFrom           string
	PublicBaseURL      string
	CookieSecure       bool
	CORSAllowedOrigins []string
	GoogleClientIDs    []string
	TrustedProxyCIDRs  []netip.Prefix
}

// Load obtiene la configuración desde el entorno y falla antes de abrir puertos
// o conexiones cuando falta un valor obligatorio.
func Load() (Config, error) {
	return load(os.Getenv)
}

func load(getenv func(string) string) (Config, error) {
	databaseURL := getenv(databaseURLEnv)
	if databaseURL == "" {
		return Config{}, fmt.Errorf("%s debe estar definido", databaseURLEnv)
	}

	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		return Config{}, fmt.Errorf("%s no es una URL válida: %w", databaseURLEnv, err)
	}
	if (parsedURL.Scheme != "postgres" && parsedURL.Scheme != "postgresql") || parsedURL.Host == "" {
		return Config{}, fmt.Errorf("%s debe usar una URL PostgreSQL con host", databaseURLEnv)
	}

	httpAddr := getenv(httpAddrEnv)
	if httpAddr == "" {
		return Config{}, fmt.Errorf("%s debe estar definido", httpAddrEnv)
	}
	smtpAddr := getenv(smtpAddrEnv)
	if smtpAddr == "" {
		return Config{}, fmt.Errorf("%s debe estar definido", smtpAddrEnv)
	}
	smtpFrom := getenv(smtpFromEnv)
	if smtpFrom == "" {
		return Config{}, fmt.Errorf("%s debe estar definido", smtpFromEnv)
	}
	publicBaseURL := getenv(publicBaseURLEnv)
	parsedPublicURL, err := url.Parse(publicBaseURL)
	if err != nil || !validPublicBaseURL(parsedPublicURL) {
		return Config{}, fmt.Errorf("%s debe ser una URL absoluta válida", publicBaseURLEnv)
	}
	corsAllowedOrigins, err := parseAllowedOrigins(getenv(corsAllowedOriginsEnv))
	if err != nil {
		return Config{}, err
	}
	googleClientIDs := parseCommaSeparated(getenv(googleClientIDsEnv))
	trustedProxyCIDRs, err := parseTrustedProxyCIDRs(getenv(trustedProxyCIDRsEnv))
	if err != nil {
		return Config{}, err
	}

	return Config{
		DatabaseURL:        databaseURL,
		HTTPAddr:           httpAddr,
		SMTPAddr:           smtpAddr,
		SMTPFrom:           smtpFrom,
		PublicBaseURL:      publicBaseURL,
		CookieSecure:       parsedPublicURL.Scheme == "https",
		CORSAllowedOrigins: corsAllowedOrigins,
		GoogleClientIDs:    googleClientIDs,
		TrustedProxyCIDRs:  trustedProxyCIDRs,
	}, nil
}

func parseTrustedProxyCIDRs(raw string) ([]netip.Prefix, error) {
	values := parseCommaSeparated(raw)
	prefixes := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			return nil, fmt.Errorf("%s contiene un CIDR inválido: %q", trustedProxyCIDRsEnv, value)
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	return prefixes, nil
}

func parseCommaSeparated(raw string) []string {
	var values []string
	for _, value := range strings.Split(raw, ",") {
		if value = strings.TrimSpace(value); value != "" {
			values = append(values, value)
		}
	}
	return values
}

func validPublicBaseURL(parsedURL *url.URL) bool {
	if parsedURL.Host == "" {
		return false
	}
	if parsedURL.Scheme == "https" {
		return true
	}
	if parsedURL.Scheme != "http" {
		return false
	}
	hostname := strings.ToLower(parsedURL.Hostname())
	return hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1"
}

func parseAllowedOrigins(raw string) ([]string, error) {
	if raw == "" {
		return nil, fmt.Errorf("%s debe estar definido", corsAllowedOriginsEnv)
	}

	origins := make([]string, 0, len(strings.Split(raw, ",")))
	seen := make(map[string]struct{})
	for _, value := range strings.Split(raw, ",") {
		origin := strings.TrimSpace(value)
		parsed, err := url.Parse(origin)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return nil, fmt.Errorf("%s contiene un origen inválido: %q", corsAllowedOriginsEnv, origin)
		}
		normalized := strings.ToLower(parsed.Scheme) + "://" + strings.ToLower(parsed.Host)
		if _, duplicate := seen[normalized]; duplicate {
			continue
		}
		seen[normalized] = struct{}{}
		origins = append(origins, normalized)
	}
	if len(origins) == 0 {
		return nil, fmt.Errorf("%s debe contener al menos un origen", corsAllowedOriginsEnv)
	}
	return origins, nil
}
