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
