// Package observability configures technical telemetry at the application's edge.
package observability

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

const serviceName = "tournaments-manager-api"

// ConfigureLogging installs the safe production logger independently from
// tracing so one-shot commands keep the same structured log contract.
func ConfigureLogging() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
}

// Configure makes JSON logging the default and installs a trace provider. An
// empty endpoint retains local logging and metrics without exporting traces.
func Configure(ctx context.Context, tracesEndpoint string) (func(context.Context) error, error) {
	ConfigureLogging()

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
