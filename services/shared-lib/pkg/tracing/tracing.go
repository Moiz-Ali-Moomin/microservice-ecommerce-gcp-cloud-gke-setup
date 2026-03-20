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
//   - OTEL_REQUIRE_TLS            – set to "true" for external/cross-cluster collectors
//   - OTEL_INSECURE               – legacy alias, set to "true" for explicit plaintext
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

	// RUNTIME FIX: The in-cluster OTel Collector serves plaintext gRPC on 4317.
	// The old code ALWAYS used TLS unless OTEL_INSECURE=true, meaning every
	// service had a TLS handshake failure against the plaintext collector and
	// ALL traces were silently dropped.
	//
	// New logic:
	//  - Default: plaintext (correct for in-cluster with mTLS handled by Istio)
	//  - OTEL_REQUIRE_TLS=true: enable TLS (for external ingestion endpoints)
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
	}
	if os.Getenv("OTEL_REQUIRE_TLS") == "true" {
		opts = append(opts, otlptracegrpc.WithTLSCredentials(
			credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12}),
		))
	} else {
		// In-cluster: Istio mTLS handles transport security at the sidecar level.
		// The OTel collector gRPC listener is plaintext within the mesh.
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(getEnv("SERVICE_VERSION", "unknown")),
			semconv.DeploymentEnvironment(getEnv("ENV", "production")),
		),
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
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
