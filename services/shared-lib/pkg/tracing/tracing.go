package tracing

import (
	"context"
	"crypto/tls"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc/credentials"
)

// InitTracer initialises the global OpenTelemetry tracer provider.
//
// Environment variables consumed:
//   - OTEL_EXPORTER_OTLP_ENDPOINT – gRPC endpoint of the OTel collector (required)
//   - OTEL_INSECURE               – set to "true" ONLY in local/dev environments
//   - SERVICE_VERSION             – semantic version string, e.g. "2.1.0"
//   - BUILD_SHA                   – git commit SHA injected at build time
//   - ENV                         – deployment environment (dev/staging/prod)
func InitTracer(serviceName string) (*sdktrace.TracerProvider, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "otel-collector.platform-observability.svc.cluster.local:4317"
	}

	// Build gRPC dial options. TLS is ON by default; only disabled when OTEL_INSECURE=true.
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
	}
	if os.Getenv("OTEL_INSECURE") == "true" {
		opts = append(opts, otlptracegrpc.WithInsecure())
	} else {
		opts = append(opts, otlptracegrpc.WithTLSCredentials(
			credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12}),
		))
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	// Enrich trace resource with service metadata for better attribution in Tempo/Jaeger.
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(getEnv("SERVICE_VERSION", "unknown")),
			semconv.DeploymentEnvironment(getEnv("ENV", "production")),
		),
		resource.WithFromEnv(),   // Picks up OTEL_RESOURCE_ATTRIBUTES if set
		resource.WithProcess(),   // Adds PID, executable name
		resource.WithOS(),        // Adds OS type
		resource.WithContainer(), // Adds container ID (useful in K8s)
		resource.WithHost(),      // Adds hostname / node name
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		// Sample 100% of traces in dev, 10% in prod (override via OTEL_TRACES_SAMPLER env var)
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(getSampleRate()))),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getSampleRate() float64 {
	switch os.Getenv("ENV") {
	case "prod", "production":
		return 0.1 // 10% sampling in production
	case "staging":
		return 0.5
	default:
		return 1.0 // 100% in dev/local
	}
}
