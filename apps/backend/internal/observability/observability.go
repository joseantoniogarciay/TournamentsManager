// Package observability configures technical telemetry at the application's edge.
package observability

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

const serviceName = "tournaments-manager-api"

// Configure makes JSON logging the default and installs a trace provider. An
// empty endpoint retains local logging and metrics without exporting traces.
func Configure(ctx context.Context, tracesEndpoint string) (func(context.Context) error, error) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	resource, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(serviceName)),
	)
	if err != nil {
		return nil, err
	}
	options := []sdktrace.TracerProviderOption{sdktrace.WithResource(resource)}
	if tracesEndpoint != "" {
		exporterOptions := []otlptracehttp.Option{otlptracehttp.WithEndpointURL(tracesEndpoint)}
		if strings.HasPrefix(tracesEndpoint, "http://") {
			exporterOptions = append(exporterOptions, otlptracehttp.WithInsecure())
		}
		exporter, err := otlptracehttp.New(ctx, exporterOptions...)
		if err != nil {
			return nil, err
		}
		options = append(options, sdktrace.WithBatcher(exporter))
	}
	provider := sdktrace.NewTracerProvider(options...)
	otel.SetTracerProvider(provider)

	return func(shutdownContext context.Context) error {
		return provider.Shutdown(shutdownContext)
	}, nil
}

// HTTPHandler adds OpenTelemetry spans, Prometheus metrics and one safe JSON
// log per completed request. It intentionally records route templates, never
// request bodies, queries, cookies, tokens, account IDs, or query strings.
func HTTPHandler(next http.Handler) http.Handler {
	metrics := newHTTPMetrics()
	application := otelhttp.NewHandler(
		metrics.wrap(requestLogger(restoreRequestURL(next))),
		"HTTP server",
	)
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/metrics" {
			MetricsHandler().ServeHTTP(writer, request)
			return
		}
		// otelhttp records the complete URL as a semantic attribute. Give it a
		// copy without its query string, while restoreRequestURL later gives the
		// actual handler the original URL and behavior.
		originalURL := *request.URL
		observedRequest := request.Clone(context.WithValue(request.Context(), originalURLContextKey{}, &originalURL))
		observedRequest.URL.RawQuery = ""
		application.ServeHTTP(writer, observedRequest)
	})
}

type originalURLContextKey struct{}

func restoreRequestURL(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		originalURL, ok := request.Context().Value(originalURLContextKey{}).(*url.URL)
		if !ok {
			next.ServeHTTP(writer, request)
			return
		}
		applicationRequest := request.Clone(request.Context())
		applicationRequest.URL = originalURL
		next.ServeHTTP(writer, applicationRequest)
		request.Pattern = applicationRequest.Pattern
	})
}

// MetricsHandler exposes only aggregated Prometheus measurements.
func MetricsHandler() http.Handler { return promhttp.Handler() }

type httpMetrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

func newHTTPMetrics() httpMetrics {
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "tournaments_manager",
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "Completed HTTP requests.",
	}, []string{"method", "route", "status"})
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "tournaments_manager",
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Help:      "Completed HTTP request duration.",
	}, []string{"method", "route", "status"})
	prometheus.MustRegister(requests, duration)
	return httpMetrics{requests: requests, duration: duration}
}

func (m httpMetrics) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: writer, status: http.StatusOK}
		next.ServeHTTP(recorder, request)
		route := request.Pattern
		if route == "" {
			route = "unmatched"
		}
		status := strconv.Itoa(recorder.status)
		labels := prometheus.Labels{"method": request.Method, "route": route, "status": status}
		m.requests.With(labels).Inc()
		m.duration.With(labels).Observe(time.Since(started).Seconds())
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: writer, status: http.StatusOK}
		next.ServeHTTP(recorder, request)
		spanContext := trace.SpanFromContext(request.Context()).SpanContext()
		attributes := []any{
			"method", request.Method,
			"route", routeName(request.Pattern),
			"status", recorder.status,
			"duration_ms", time.Since(started).Milliseconds(),
		}
		if spanContext.IsValid() {
			attributes = append(attributes, "trace_id", spanContext.TraceID().String(), "span_id", spanContext.SpanID().String())
		}
		slog.Info("HTTP request completed", attributes...)
	})
}

func routeName(pattern string) string {
	if pattern == "" {
		return "unmatched"
	}
	return pattern
}

// QueryTracer creates child spans for PostgreSQL queries without recording SQL
// text or arguments, which can contain personal data and credentials.
type QueryTracer struct{}

// TraceQueryStart starts a span without copying SQL text or arguments into it.
func (QueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	ctx, _ = otel.Tracer(serviceName+"/postgres").Start(ctx, "postgresql.query",
		trace.WithAttributes(spanAttributes(queryOperation(data.SQL))...),
	)
	return ctx
}

// TraceQueryEnd records a safe failure summary and completes the query span.
func (QueryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	span := trace.SpanFromContext(ctx)
	if data.Err != nil {
		span.RecordError(data.Err)
		span.SetStatus(codes.Error, "database query failed")
	}
	span.End()
}

func queryOperation(sql string) string {
	fields := strings.Fields(sql)
	if len(fields) == 0 {
		return "unknown"
	}
	return strings.ToUpper(fields[0])
}

func spanAttributes(operation string) []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.DBSystemNamePostgreSQL,
		semconv.DBOperationName(operation),
	}
}
