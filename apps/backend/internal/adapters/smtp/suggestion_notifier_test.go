package smtp

import (
	"strings"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/suggestions"
)

func TestSuggestionMessageKeepsUserContentOutOfHeaders(t *testing.T) {
	t.Parallel()
	mailer := Mailer{from: "FastTourney <no-reply@example.test>", subjectPrefix: "[DEV]"}
	notifier, err := NewSuggestionNotifier(mailer, "owner@example.test")
	if err != nil {
		t.Fatalf("NewSuggestionNotifier() error = %v", err)
	}
	item := suggestions.Item{ID: "019abcde-1111-7111-8111-111111111111", Username: "person", Body: "Primera línea\nSubject: contenido", CreatedAt: "2026-09-13T10:00:00Z"}
	message := string(suggestionMessage(notifier.recipient, notifier.mailer.from, notifier.mailer.subjectPrefix, item))
	headers, body, _ := strings.Cut(message, "\r\n\r\n")
	if strings.Contains(headers, item.Body) || strings.Contains(headers, "person") {
		t.Fatalf("headers contain user content: %q", headers)
	}
	for _, expected := range []string{"@person", item.ID, item.CreatedAt, item.Body} {
		if !strings.Contains(body, expected) {
			t.Errorf("body does not contain %q", expected)
		}
	}
}

func TestNewSuggestionNotifierRejectsHeaderInjection(t *testing.T) {
	t.Parallel()
	if _, err := NewSuggestionNotifier(Mailer{}, "owner@example.test\r\nBcc: attacker@example.test"); err == nil {
		t.Fatal("NewSuggestionNotifier() error = nil, want invalid recipient")
	}
}
