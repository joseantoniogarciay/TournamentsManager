package observability

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

// QueryTracer creates child spans for PostgreSQL queries without recording SQL
// text or arguments, which can contain personal data and credentials.
type QueryTracer struct{}

func (QueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	queryName, operation := queryMetadata(data.SQL)
	spanName := "postgresql.query"
	attributes := spanAttributes(operation)
	if queryName != "" {
		spanName = "postgresql." + queryName
		attributes = append(attributes, attribute.String("db.query.name", queryName))
	}
	ctx, _ = otel.Tracer(serviceName+"/postgres").Start(ctx, spanName, trace.WithAttributes(attributes...))
	return ctx
}

func (QueryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	span := trace.SpanFromContext(ctx)
	if data.Err != nil {
		span.RecordError(data.Err)
		span.SetStatus(codes.Error, "database query failed")
	}
	span.End()
}

func queryMetadata(sql string) (name, operation string) {
	for _, line := range strings.Split(sql, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "-- name:") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				name = fields[2]
			}
			continue
		}
		if strings.HasPrefix(line, "--") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			return name, strings.ToUpper(fields[0])
		}
	}
	return name, "unknown"
}

func spanAttributes(operation string) []attribute.KeyValue {
	return []attribute.KeyValue{semconv.DBSystemNamePostgreSQL, semconv.DBOperationName(operation)}
}
