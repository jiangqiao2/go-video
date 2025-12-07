package observability

import (
	"context"
	"os"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// TracerConfig defines how to connect to OTLP and sample traces.
type TracerConfig struct {
	ServiceName string
	Endpoint    string
	Environment string
	SampleRatio float64
	Insecure    bool
}

// DefaultTracerConfig derives config from env vars with sensible defaults.
func DefaultTracerConfig(serviceName string) TracerConfig {
	sample := 0.1
	if v := os.Getenv("OTEL_TRACES_SAMPLING_RATE"); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil && parsed >= 0 && parsed <= 1 {
			sample = parsed
		}
	}

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	env := firstNonEmpty(os.Getenv("ENV"), os.Getenv("APP_ENV"), "dev")
	insecure := strings.EqualFold(os.Getenv("OTEL_EXPORTER_OTLP_INSECURE"), "true") ||
		strings.EqualFold(os.Getenv("OTEL_EXPORTER_OTLP_INSECURE"), "1") ||
		endpointHasLocalhost(endpoint)

	return TracerConfig{
		ServiceName: serviceName,
		Endpoint:    endpoint,
		Environment: env,
		SampleRatio: sample,
		Insecure:    insecure,
	}
}

// InitTracer installs a global tracer provider and propagator; returns a shutdown hook.
func InitTracer(ctx context.Context, cfg TracerConfig) (func(context.Context) error, error) {
	if cfg.Endpoint == "" {
		return nil, nil
	}

	sample := cfg.SampleRatio
	if sample <= 0 || sample > 1 {
		sample = 0.1
	}

	clientOpts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(cfg.Endpoint)}
	if cfg.Insecure {
		clientOpts = append(clientOpts, otlptracegrpc.WithInsecure())
	}

	exporter, err := otlptrace.New(ctx, otlptracegrpc.NewClient(clientOpts...))
	if err != nil {
		return nil, err
	}

	res, err := sdkresource.New(ctx,
		sdkresource.WithFromEnv(),
		sdkresource.WithProcess(),
		sdkresource.WithHost(),
		sdkresource.WithTelemetrySDK(),
		sdkresource.WithAttributes(
			attribute.String("service.name", cfg.ServiceName),
			attribute.String("service.version", "1.0.0"),
			attribute.String("deployment.environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(sample))),
		sdktrace.WithBatcher(exporter),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func endpointHasLocalhost(endpoint string) bool {
	if endpoint == "" {
		return false
	}
	lower := strings.ToLower(endpoint)
	return strings.Contains(lower, "127.0.0.1") || strings.Contains(lower, "localhost")
}
