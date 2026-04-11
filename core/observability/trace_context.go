package observability

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TraceContext handles propagation of trace context across HTTP requests.
type TraceContext struct {
	propagator propagation.TextMapPropagator
}

// NewTraceContext creates a new TraceContext using the global propagator.
func NewTraceContext() *TraceContext {
	return &TraceContext{
		propagator: otel.GetTextMapPropagator(),
	}
}

// Inject adds trace context to the outgoing HTTP request.
func (tc *TraceContext) Inject(ctx context.Context, req *http.Request) {
	if tc.propagator != nil {
		tc.propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))
	}
}

// Extract retrieves trace context from the incoming HTTP request.
func (tc *TraceContext) Extract(req *http.Request) context.Context {
	if tc.propagator != nil {
		return tc.propagator.Extract(req.Context(), propagation.HeaderCarrier(req.Header))
	}
	return req.Context()
}

// Middleware returns an HTTP middleware that extracts trace context from incoming requests.
func (tc *TraceContext) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := tc.Extract(r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TraceIDFromSpan returns the trace ID from the current span context.
func TraceIDFromSpan(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// SpanIDFromSpan returns the span ID from the current span context.
func SpanIDFromSpan(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// NewTraceID generates a new trace ID for testing purposes.
// In production, trace IDs are generated automatically by OpenTelemetry.
func NewTraceID() string {
	tp := otel.GetTracerProvider()
	if tp == nil {
		return ""
	}
	tracer := tp.Tracer("brain.daemon.tracecontext")
	if tracer == nil {
		return ""
	}
	// Create a dummy span to get a trace ID
	_, span := tracer.Start(context.Background(), "trace-id-generation")
	traceID := span.SpanContext().TraceID().String()
	span.End()
	return traceID
}
