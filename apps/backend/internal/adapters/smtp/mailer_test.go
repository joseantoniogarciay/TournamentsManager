package smtp

import (
	"strings"
	"testing"
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

	message, err := verificationMessage("person@example.test", "no-reply@example.test", "https://links.example.test/link/confirm?token=test-token")
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
	} {
		if !strings.Contains(contents, expected) {
			t.Errorf("message does not contain %q", expected)
		}
	}
	if strings.Contains(contents, "/verify-registration") {
		t.Error("message still uses the old verification route")
	}
}
