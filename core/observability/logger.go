package observability

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// LoggerConfig holds configuration for the structured logger.
type LoggerConfig struct {
	// Level is the minimum log level (debug, info, warn, error).
	Level string
	// Format is the output format (json, text).
	Format string
	// Output is the writer for log output (defaults to stderr).
	Output io.Writer
	// ServiceName is added to all log records.
	ServiceName string
	// Version is the service version added to all log records.
	Version string
}

// NewLogger creates a structured slog.Logger based on the provided configuration.
// If cfg is nil, defaults to JSON format, info level, and stderr output.
func NewLogger(cfg *LoggerConfig) *slog.Logger {
	if cfg == nil {
		cfg = &LoggerConfig{
			Level:  "info",
			Format: "json",
			Output: os.Stderr,
		}
	}
	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}

	level := parseLevel(cfg.Level)
	opts := &slog.HandlerOptions{
		Level: level,
		AddSource: true,
	}

	var handler slog.Handler
	switch cfg.Format {
	case "text":
		handler = slog.NewTextHandler(cfg.Output, opts)
	default:
		handler = slog.NewJSONHandler(cfg.Output, opts)
	}

	// Add common attributes
	attrs := []slog.Attr{
		slog.String("service", cfg.ServiceName),
	}
	if cfg.Version != "" {
		attrs = append(attrs, slog.String("version", cfg.Version))
	}

	logger := slog.New(handler.WithAttrs(attrs))
	slog.SetDefault(logger)
	return logger
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Common log attribute keys for Brain daemon operations.
const (
	AttrArtifactKind   = "artifact.kind"
	AttrArtifactID     = "artifact.id"
	AttrScope          = "artifact.scope"
	AttrPolicyClass    = "policy.class"
	AttrModelID        = "model.id"
	AttrCapabilityTier = "model.capability_tier"
	AttrTraceID        = "trace.id"
	AttrExecutionID   = "execution.id"
	AttrComponent      = "component"
	AttrError          = "error"
)

// WithTraceID adds a trace ID to the context for log correlation.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, AttrTraceID, traceID)
}

// TraceIDFromContext retrieves the trace ID from context, or returns "unknown".
func TraceIDFromContext(ctx context.Context) string {
	if tid, ok := ctx.Value(AttrTraceID).(string); ok {
		return tid
	}
	return "unknown"
}

// WithExecutionID adds an execution ID to the context.
func WithExecutionID(ctx context.Context, execID string) context.Context {
	return context.WithValue(ctx, AttrExecutionID, execID)
}

// ExecutionIDFromContext retrieves the execution ID from context.
func ExecutionIDFromContext(ctx context.Context) string {
	if eid, ok := ctx.Value(AttrExecutionID).(string); ok {
		return eid
	}
	return ""
}
