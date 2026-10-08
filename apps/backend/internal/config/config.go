// Package config loads and validates API runtime configuration.
package config

import (
	"fmt"
	"net/mail"
	"net/netip"
	"net/url"
	"os"
	"regexp"
	"strings"
)

const (
	databaseURLEnv         = "DATABASE_URL"
	httpAddrEnv            = "HTTP_ADDR"
	smtpAddrEnv            = "SMTP_ADDR"
	smtpFromEnv            = "SMTP_FROM"
	smtpUsernameEnv        = "SMTP_USERNAME"
	smtpPasswordEnv        = "SMTP_PASSWORD"
	emailSubjectPrefixEnv  = "EMAIL_SUBJECT_PREFIX"
	suggestionRecipientEnv = "SUGGESTION_RECIPIENT"
	publicBaseURLEnv       = "PUBLIC_BASE_URL"
	corsAllowedOriginsEnv  = "CORS_ALLOWED_ORIGINS"
	googleClientIDsEnv     = "GOOGLE_CLIENT_IDS"
	trustedProxyCIDRsEnv   = "TRUSTED_PROXY_CIDRS"
	edgeProxyAuthTokenEnv  = "EDGE_PROXY_AUTH_TOKEN" // #nosec G101 -- runtime environment variable name, not a credential.
	otelTracesEndpointEnv  = "OTEL_TRACES_ENDPOINT"
)

// AppleConfig is inactive until every real server-side Apple value is supplied.
type AppleConfig struct {
	ServiceID, TeamID, KeyID, PrivateKeyFile, RedirectURI, NativeScheme string
	Enabled                                                             bool
}

// Config contains only the configuration needed to start the API.
type Config struct {
	DatabaseURL         string
	HTTPAddr            string
	SMTPAddr            string
	SMTPFrom            string
	SMTPUsername        string
	SMTPPassword        string
	EmailSubjectPrefix  string
	SuggestionRecipient string
	PublicBaseURL       string
	CookieSecure        bool
	CORSAllowedOrigins  []string
	GoogleClientIDs     []string
	Apple               AppleConfig
	TrustedProxyCIDRs   []netip.Prefix
	EdgeProxyAuthToken  string
	OTELTracesEndpoint  string
}

// Load gets configuration from the environment and fails before opening ports
// or connections when a required value is missing.
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
	smtpUsername := getenv(smtpUsernameEnv)
	smtpPassword := getenv(smtpPasswordEnv)
	if (smtpUsername == "") != (smtpPassword == "") {
		return Config{}, fmt.Errorf("%s y %s deben definirse juntos", smtpUsernameEnv, smtpPasswordEnv)
	}
	emailSubjectPrefix := getenv(emailSubjectPrefixEnv)
	if strings.ContainsAny(emailSubjectPrefix, "\r\n") {
		return Config{}, fmt.Errorf("%s no puede contener saltos de línea", emailSubjectPrefixEnv)
	}
	suggestionRecipient := strings.TrimSpace(getenv(suggestionRecipientEnv))
	if suggestionRecipient != "" {
		parsedRecipient, err := mail.ParseAddress(suggestionRecipient)
		if err != nil || parsedRecipient.Address == "" || strings.ContainsAny(suggestionRecipient, "\r\n") {
			return Config{}, fmt.Errorf("%s debe contener un email válido", suggestionRecipientEnv)
		}
		suggestionRecipient = parsedRecipient.Address
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
	filtered := googleClientIDs[:0]
	for _, id := range googleClientIDs {
		if configuredValue(id) {
			filtered = append(filtered, id)
		}
	}
	googleClientIDs = filtered
	apple, err := loadApple(getenv)
	if err != nil {
		return Config{}, err
	}
	trustedProxyCIDRs, err := parseTrustedProxyCIDRs(getenv(trustedProxyCIDRsEnv))
	if err != nil {
		return Config{}, err
	}
	edgeProxyAuthToken := getenv(edgeProxyAuthTokenEnv)
	if strings.ContainsAny(edgeProxyAuthToken, "\r\n") {
		return Config{}, fmt.Errorf("%s no puede contener saltos de línea", edgeProxyAuthTokenEnv)
	}
	otelTracesEndpoint, err := parseOTELTracesEndpoint(getenv(otelTracesEndpointEnv))
	if err != nil {
		return Config{}, err
	}

	return Config{
		DatabaseURL:         databaseURL,
		HTTPAddr:            httpAddr,
		SMTPAddr:            smtpAddr,
		SMTPFrom:            smtpFrom,
		SMTPUsername:        smtpUsername,
		SMTPPassword:        smtpPassword,
		EmailSubjectPrefix:  emailSubjectPrefix,
		SuggestionRecipient: suggestionRecipient,
		PublicBaseURL:       publicBaseURL,
		CookieSecure:        parsedPublicURL.Scheme == "https",
		CORSAllowedOrigins:  corsAllowedOrigins,
		GoogleClientIDs:     googleClientIDs,
		Apple:               apple,
		TrustedProxyCIDRs:   trustedProxyCIDRs,
		EdgeProxyAuthToken:  edgeProxyAuthToken,
		OTELTracesEndpoint:  otelTracesEndpoint,
	}, nil
}

func parseOTELTracesEndpoint(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("%s debe ser una URL HTTP(S) absoluta sin credenciales ni query", otelTracesEndpointEnv)
	}
	return parsed.String(), nil
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

// configuredValue ensures example placeholders never enable a live provider.
func configuredValue(value string) bool {
	value = strings.TrimSpace(value)
	upper := strings.ToUpper(value)
	return value != "" && !strings.Contains(upper, "PLACEHOLDER") && !strings.Contains(upper, "REPLACE_") && !strings.Contains(upper, "REPLACE-WITH-") && !strings.Contains(upper, "CHANGE-ME") && !strings.ContainsAny(value, "<>")
}

func loadApple(getenv func(string) string) (AppleConfig, error) {
	apple := AppleConfig{ServiceID: strings.TrimSpace(getenv("APPLE_SERVICE_ID")), TeamID: strings.TrimSpace(getenv("APPLE_TEAM_ID")), KeyID: strings.TrimSpace(getenv("APPLE_KEY_ID")), PrivateKeyFile: strings.TrimSpace(getenv("APPLE_PRIVATE_KEY_FILE")), RedirectURI: strings.TrimSpace(getenv("APPLE_REDIRECT_URI")), NativeScheme: strings.TrimSpace(getenv("APPLE_NATIVE_SCHEME"))}
	for _, value := range []string{apple.ServiceID, apple.TeamID, apple.KeyID, apple.PrivateKeyFile, apple.RedirectURI, apple.NativeScheme} {
		if !configuredValue(value) {
			return apple, nil
		}
	}
	parsed, err := url.Parse(apple.RedirectURI)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.Path != "/v1/apple-callback" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return AppleConfig{}, fmt.Errorf("APPLE_REDIRECT_URI debe ser un callback HTTPS /v1/apple-callback")
	}
	if apple.NativeScheme != "fasttourney-dev" && apple.NativeScheme != "fasttourney" {
		return AppleConfig{}, fmt.Errorf("APPLE_NATIVE_SCHEME debe ser la variante dev o prod")
	}
	if !regexp.MustCompile(`^[A-Z0-9]{10}$`).MatchString(apple.TeamID) || !regexp.MustCompile(`^[A-Z0-9]{10}$`).MatchString(apple.KeyID) || !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.-]+$`).MatchString(apple.ServiceID) {
		return AppleConfig{}, fmt.Errorf("identificadores Apple no válidos")
	}
	apple.Enabled = true
	return apple, nil
}
