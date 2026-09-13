package smtp

import (
	"context"
	"fmt"
	"mime"
	"net/mail"
	"strings"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/suggestions"
)

// SuggestionNotifier sends a secondary owner alert through the configured mailer.
type SuggestionNotifier struct {
	mailer    Mailer
	recipient string
}

// NewSuggestionNotifier validates the environment-specific recipient.
func NewSuggestionNotifier(mailer Mailer, recipient string) (SuggestionNotifier, error) {
	parsed, err := mail.ParseAddress(recipient)
	if err != nil || parsed.Address == "" || strings.ContainsAny(recipient, "\r\n") {
		return SuggestionNotifier{}, fmt.Errorf("destinatario de sugerencias inválido")
	}
	return SuggestionNotifier{mailer: mailer, recipient: parsed.Address}, nil
}

// NotifySuggestion sends user content only in the MIME body, never in headers.
func (n SuggestionNotifier) NotifySuggestion(ctx context.Context, item suggestions.Item) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	message := suggestionMessage(n.recipient, n.mailer.from, n.mailer.subjectPrefix, item)
	if err := n.mailer.send([]string{n.recipient}, message); err != nil {
		return fmt.Errorf("SMTP: %w", err)
	}
	return nil
}

func suggestionMessage(recipient, from, subjectPrefix string, item suggestions.Item) []byte {
	return []byte(
		"To: " + recipient + "\r\n" +
			"From: " + from + "\r\n" +
			"Subject: " + mime.QEncoding.Encode("UTF-8", prefixedSubject(subjectPrefix, "Nueva sugerencia de FastTourney")) + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
			"Usuario: @" + item.Username + "\r\n" +
			"Sugerencia: " + item.ID + "\r\n" +
			"Fecha: " + item.CreatedAt + "\r\n\r\n" +
			item.Body + "\r\n",
	)
}
