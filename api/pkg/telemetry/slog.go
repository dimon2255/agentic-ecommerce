package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// TracedHandler wraps a slog.Handler to inject trace_id and span_id from OpenTelemetry context.
type TracedHandler struct {
	inner slog.Handler
}

// NewTracedHandler creates a slog handler that injects trace context into log records.
func NewTracedHandler(inner slog.Handler) *TracedHandler {
	return &TracedHandler{inner: inner}
}

func (h *TracedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *TracedHandler) Handle(ctx context.Context, record slog.Record) error {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		record.AddAttrs(
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
	}
	return h.inner.Handle(ctx, record)
}

func (h *TracedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TracedHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *TracedHandler) WithGroup(name string) slog.Handler {
	return &TracedHandler{inner: h.inner.WithGroup(name)}
}
