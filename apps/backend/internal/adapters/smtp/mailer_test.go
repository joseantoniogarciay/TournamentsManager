package smtp

import (
	"mime"
	"strings"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
)

func TestNewMailerRequiresHTTPSOutsideLoopback(t *testing.T) {
	t.Parallel()

	if _, err := NewMailer("127.0.0.1:1025", "no-reply@example.test", "http://example.test"); err == nil {
		t.Fatal("NewMailer() error = nil, want invalid non-HTTPS public URL")
	}
	if _, err := NewMailer("127.0.0.1:1025", "no-reply@example.test", "https://links.example.test"); err != nil {
		t.Fatalf("NewMailer() error = %v", err)
	}
}

func TestVerificationMessageUsesConfirmationRouteAndMultipartDesign(t *testing.T) {
	t.Parallel()

	message, err := verificationMessage(registration.LocaleSpanish, "person@example.test", "no-reply@example.test", "https://links.example.test/link/confirm?token=test-token")
	if err != nil {
		t.Fatalf("verificationMessage() error = %v", err)
	}
	contents := string(message)
	for _, expected := range []string{
		"multipart/alternative",
		"text/plain; charset=UTF-8",
		"text/html; charset=UTF-8",
		"https://links.example.test/link/confirm?token=test-token",
		"Fast Tourney",
		"background:#155eef;background-image:linear-gradient(135deg,transparent 0%,transparent 35%,#7f56d9 100%)",
	} {
		if !strings.Contains(contents, expected) {
			t.Errorf("message does not contain %q", expected)
		}
	}
	if strings.Contains(contents, "/verify-registration") {
		t.Error("message still uses the old verification route")
	}
	if strings.Contains(contents, "#5b4bff") || strings.Contains(contents, "#e84a8a") {
		t.Error("message still uses the previous violet-pink palette")
	}
}

func TestVerificationMessageLocalizesAllSupportedLocales(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		locale   registration.Locale
		subject  string
		language string
	}{
		{registration.LocaleSpanish, "Verifica tu cuenta de Fast Tourney", `lang="es"`},
		{registration.LocaleEnglish, "Verify your Fast Tourney account", `lang="en"`},
		{registration.LocaleItalian, "Verifica il tuo account Fast Tourney", `lang="it"`},
		{registration.LocaleFrench, "Vérifiez votre compte Fast Tourney", `lang="fr"`},
	} {
		t.Run(string(test.locale), func(t *testing.T) {
			t.Parallel()

			message, err := verificationMessage(test.locale, "person@example.test", "no-reply@example.test", "https://links.example.test/link/confirm?token=test-token")
			if err != nil {
				t.Fatalf("verificationMessage() error = %v", err)
			}
			contents := string(message)
			expectedSubject := mime.QEncoding.Encode("UTF-8", test.subject)
			if !strings.Contains(contents, expectedSubject) {
				t.Errorf("message does not contain localized subject %q", expectedSubject)
			}
			if !strings.Contains(contents, test.language) {
				t.Errorf("message does not contain localized HTML language %q", test.language)
			}
		})
	}
}

func TestVerificationMessageRejectsUnsupportedLocale(t *testing.T) {
	t.Parallel()

	if _, err := verificationMessage("de", "person@example.test", "no-reply@example.test", "https://links.example.test/link/confirm?token=test-token"); err == nil {
		t.Fatal("verificationMessage() error = nil, want unsupported locale error")
	}
}
