package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// TracerConfig holds configuration for the OpenTelemetry tracer.
type TracerConfig struct {
	// ServiceName is the name of the service (e.g., "brain-daemon").
	ServiceName string
	// ServiceVersion is the version of the service.
	ServiceVersion string
	// OTLPEndpoint is the OTLP collector endpoint (e.g., "localhost:4318").
	OTLPEndpoint string
	// OTLPInsecure disables TLS for the OTLP endpoint.
	OTLPInsecure bool
	// SampleRate is the sampling rate (0.0-1.0, 1.0 = sample all).
	SampleRate float64
	// Enabled enables/disables tracing.
	Enabled bool
}

// DefaultTracerConfig returns a TracerConfig with sensible defaults.
func DefaultTracerConfig() *TracerConfig {
	return &TracerConfig{
		ServiceName:    "brain-daemon",
		ServiceVersion: "0.1.0",
		OTLPEndpoint:   "localhost:4318",
		OTLPInsecure:   true,
		SampleRate:     1.0,
		Enabled:        false, // Disabled by default, must be explicitly enabled
	}
}

// InitTracer initializes the OpenTelemetry tracer provider.
// Returns the tracer provider and a shutdown function.
func InitTracer(ctx context.Context, cfg *TracerConfig) (*sdktrace.TracerProvider, func(), error) {
	if cfg == nil {
		cfg = DefaultTracerConfig()
	}

	if !cfg.Enabled {
		return nil, func() {}, nil
	}

	// Create OTLP exporter
	exporter, err := otlptrace.New(ctx, otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
		otlptracehttp.WithInsecure(),
	))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider with batching and sampling
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(cfg.SampleRate),
		)),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator for trace context
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	shutdown := func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tp.Shutdown(shutdownCtx)
	}

	return tp, shutdown, nil
}

// Tracer returns a named tracer for the given component.
func Tracer(component string) trace.Tracer {
	return otel.Tracer("brain.daemon." + component)
}

// Brain-specific span attributes for daemon operations.
var (
	AttrArtifactKindKV   = attribute.Key("brain.artifact.kind")
	AttrArtifactIDKV     = attribute.Key("brain.artifact.id")
	AttrScopeKV          = attribute.Key("brain.scope")
	AttrPolicyClassKV    = attribute.Key("brain.policy.class")
	AttrModelIDKV        = attribute.Key("brain.model.id")
	AttrCapabilityTierKV = attribute.Key("brain.model.capability_tier")
	AttrClientSurfaceKV  = attribute.Key("brain.client.surface")
	AttrSyncWorkerKV     = attribute.Key("brain.sync.worker")
)

// RecordError records an error on a span and sets the span status to error.
func RecordError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// StartSpan creates a new span with the given name and attributes.
func StartSpan(ctx context.Context, tracer trace.Tracer, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	ctx, span := tracer.Start(ctx, name)
	for _, attr := range attrs {
		span.SetAttributes(attr)
	}
	return ctx, span
}
