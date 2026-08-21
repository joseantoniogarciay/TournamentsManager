package observability

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

// HTTPHandler adds OpenTelemetry spans, Prometheus metrics and one safe JSON
// log per completed request. It intentionally records route templates, never
// request bodies, queries, cookies, tokens, account IDs, or query strings.
func HTTPHandler(next http.Handler) http.Handler {
	metrics := newHTTPMetrics()
	application := otelhttp.NewHandler(metrics.wrap(requestLogger(restoreRequestURL(next))), "HTTP server", otelhttp.WithSpanNameFormatter(func(fallback string, request *http.Request) string {
		if request.Pattern == "" {
			return fallback
		}
		return request.Pattern
	}))
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/metrics" {
			MetricsHandler().ServeHTTP(writer, request)
			return
		}
		originalURL := *request.URL
		observedContext := context.WithValue(request.Context(), originalURLContextKey{}, &originalURL)
		observedRequest := request.Clone(withEndpointFailureState(observedContext))
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
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{Namespace: "tournaments_manager", Subsystem: "http", Name: "requests_total", Help: "Completed HTTP requests."}, []string{"method", "route", "status"})
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{Namespace: "tournaments_manager", Subsystem: "http", Name: "request_duration_seconds", Help: "Completed HTTP request duration."}, []string{"method", "route", "status"})
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
		if recorder.status >= http.StatusInternalServerError && endpointFailureReason(request.Context()) == "" {
			// The handler did not classify this unexpected server response. Keep
			// the root span and its correlated log useful without guessing which
			// dependency failed; a child span may provide that narrower category.
			RecordEndpointFailure(request.Context(), "request.failed")
		}
		spanContext := trace.SpanFromContext(request.Context()).SpanContext()
		attributes := []any{"method", request.Method, "route", routeName(request.Pattern), "status", recorder.status, "duration_ms", time.Since(started).Milliseconds()}
		if reason := endpointFailureReason(request.Context()); reason != "" {
			attributes = append(attributes, "failure_reason", reason)
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
