package observability

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const failureReasonAttribute = "tournaments_manager.failure.reason"

// RecordEndpointFailure adds a safe, feature-owned reason to the HTTP root span.
// Expected rejections retain their HTTP status and do not become trace errors.
func RecordEndpointFailure(ctx context.Context, reason string) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attribute.String(failureReasonAttribute, reason))
	}
}

// RecordDatabaseEndpointFailure records the safe category of a database failure
// on the HTTP root span. QueryTracer records the same category on the child
// PostgreSQL span; keeping it on the root makes the request diagnosable without
// exporting the driver error or any query data.
func RecordDatabaseEndpointFailure(ctx context.Context, err error) {
	RecordEndpointFailure(ctx, databaseFailureReason(err))
}

func recordFailure(span trace.Span, reason, summary string) {
	span.SetAttributes(attribute.String(failureReasonAttribute, reason))
	span.SetStatus(codes.Error, summary)
}

func databaseFailureReason(err error) string {
	if reason, ok := contextFailureReason(err); ok {
		return reason
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch {
		case strings.HasPrefix(postgresError.Code, "08"):
			return "database.unavailable"
		case strings.HasPrefix(postgresError.Code, "23"):
			return "database.constraint_failed"
		}
	}
	return "database.query_failed"
}

func smtpFailureReason(err error) string {
	if reason, ok := contextFailureReason(err); ok {
		return reason
	}
	return "smtp.delivery_failed"
}

func contextFailureReason(err error) (string, bool) {
	switch {
	case errors.Is(err, context.Canceled):
		return "request.cancelled", true
	case errors.Is(err, context.DeadlineExceeded):
		return "request.timeout", true
	default:
		return "", false
	}
}
