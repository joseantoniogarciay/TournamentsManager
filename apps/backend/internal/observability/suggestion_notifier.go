package observability

import (
	"context"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/suggestions"
	"go.opentelemetry.io/otel"
)

// SuggestionNotifier measures owner notification without recording its content or identity.
type SuggestionNotifier struct{ Next suggestions.Notifier }

// NotifySuggestion creates one safe dependency span around SMTP delivery.
func (n SuggestionNotifier) NotifySuggestion(ctx context.Context, item suggestions.Item) error {
	ctx, span := otel.Tracer(serviceName+"/smtp").Start(ctx, "smtp.send.suggestion")
	defer span.End()
	err := n.Next.NotifySuggestion(ctx, item)
	if err != nil {
		recordFailure(span, smtpFailureReason(err), "SMTP delivery failed")
	}
	return err
}
