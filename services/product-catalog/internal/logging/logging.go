package logging

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/trace"
)

// New returns a slog logger that fans every record out to two sinks:
//
//   - stdout as JSON (for `docker logs`), enriched with trace_id/span_id read
//     from the active span in the context so console lines carry the same ids
//     the trace does — this is the pair the Loki derivedFields regex matches;
//   - the OTel log pipeline (OTLP → collector → Loki) via the otelslog bridge,
//     which stamps the record's own trace context from the context so a span in
//     Tempo links back to its logs.
//
// observability.Setup must run before New: the otelslog bridge binds to the
// global LoggerProvider at construction time, so the real provider has to be
// installed first or these logs never leave the process.
//
// Both correlation directions depend on call sites using the *Context variants
// (InfoContext/ErrorContext/…): the context is what carries the span.
func New(serviceName string) *slog.Logger {
	handler := &fanout{handlers: []slog.Handler{
		&traceEnrich{next: slog.NewJSONHandler(os.Stdout, nil)},
		otelslog.NewHandler(serviceName),
	}}
	return slog.New(handler).With("service", serviceName)
}

// fanout dispatches every record to all its handlers. Each handler decides its
// own level via Enabled, so fanout enables a level if any child does.
type fanout struct {
	handlers []slog.Handler
}

func (f *fanout) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range f.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle passes a clone to each handler: slog.Record shares a backing array for
// its attributes, so a handler that appends (as the enricher does) would corrupt
// the record the sibling handler sees without the copy.
func (f *fanout) Handle(ctx context.Context, r slog.Record) error {
	var firstErr error
	for _, h := range f.handlers {
		if !h.Enabled(ctx, r.Level) {
			continue
		}
		if err := h.Handle(ctx, r.Clone()); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (f *fanout) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Handler, len(f.handlers))
	for i, h := range f.handlers {
		next[i] = h.WithAttrs(attrs)
	}
	return &fanout{handlers: next}
}

func (f *fanout) WithGroup(name string) slog.Handler {
	next := make([]slog.Handler, len(f.handlers))
	for i, h := range f.handlers {
		next[i] = h.WithGroup(name)
	}
	return &fanout{handlers: next}
}

// traceEnrich adds trace_id/span_id from the active span to a record before
// delegating, so the stdout JSON carries the ids without every call site
// having to pass them.
type traceEnrich struct {
	next slog.Handler
}

func (t *traceEnrich) Enabled(ctx context.Context, level slog.Level) bool {
	return t.next.Enabled(ctx, level)
}

func (t *traceEnrich) Handle(ctx context.Context, r slog.Record) error {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return t.next.Handle(ctx, r)
}

func (t *traceEnrich) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceEnrich{next: t.next.WithAttrs(attrs)}
}

func (t *traceEnrich) WithGroup(name string) slog.Handler {
	return &traceEnrich{next: t.next.WithGroup(name)}
}
