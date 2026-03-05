package otel

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/log"
)

type OTLPHandler struct {
	logger log.Logger
}

func NewOTLPHandler(logger log.Logger) *OTLPHandler {
	return &OTLPHandler{logger: logger}
}

func (h *OTLPHandler) Enabled(
	ctx context.Context,
	level slog.Level,
) bool {
	return true
}

func (h *OTLPHandler) Handle(
	ctx context.Context,
	r slog.Record, // nolint:gocritic
) error {
	severity := levelToSeverity(r.Level)

	rec := log.Record{}
	rec.SetTimestamp(r.Time)
	rec.SetBody(log.StringValue(r.Message))
	rec.SetSeverity(severity)
	rec.SetSeverityText(r.Level.String())

	attrs := make([]log.KeyValue, 0)
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(
			attrs,
			log.String(a.Key, a.Value.String()),
		)
		return true
	})
	rec.AddAttributes(attrs...)

	h.logger.Emit(ctx, rec)
	return nil
}

func (h *OTLPHandler) WithAttrs(
	attrs []slog.Attr,
) slog.Handler {
	return &OTLPHandlerWithAttrs{
		parent: h,
		attrs:  attrs,
	}
}

func (h *OTLPHandler) WithGroup(
	name string,
) slog.Handler {
	return h
}

type OTLPHandlerWithAttrs struct {
	parent *OTLPHandler
	attrs  []slog.Attr
}

func (h *OTLPHandlerWithAttrs) Enabled(
	ctx context.Context,
	level slog.Level,
) bool {
	return h.parent.Enabled(ctx, level)
}

func (h *OTLPHandlerWithAttrs) Handle(
	ctx context.Context,
	r slog.Record, // nolint:gocritic
) error {
	severity := levelToSeverity(r.Level)

	rec := log.Record{}
	rec.SetTimestamp(r.Time)
	rec.SetBody(log.StringValue(r.Message))
	rec.SetSeverity(severity)
	rec.SetSeverityText(r.Level.String())

	attrs := make([]log.KeyValue, 0, len(h.attrs)+r.NumAttrs())
	for _, a := range h.attrs {
		attrs = append(
			attrs,
			log.String(a.Key, a.Value.String()),
		)
	}

	r.Attrs(func(a slog.Attr) bool {
		attrs = append(
			attrs,
			log.String(a.Key, a.Value.String()),
		)
		return true
	})
	rec.AddAttributes(attrs...)

	h.parent.logger.Emit(ctx, rec)
	return nil
}

func (h *OTLPHandlerWithAttrs) WithAttrs(
	attrs []slog.Attr,
) slog.Handler {
	newAttrs := make(
		[]slog.Attr,
		0,
		len(h.attrs)+len(attrs),
	)
	newAttrs = append(newAttrs, h.attrs...)
	newAttrs = append(newAttrs, attrs...)
	return &OTLPHandlerWithAttrs{
		parent: h.parent,
		attrs:  newAttrs,
	}
}

func (h *OTLPHandlerWithAttrs) WithGroup(
	name string,
) slog.Handler {
	return h
}

func levelToSeverity(level slog.Level) log.Severity {
	switch {
	case level >= slog.LevelError:
		return log.SeverityError
	case level >= slog.LevelWarn:
		return log.SeverityWarn
	case level >= slog.LevelDebug:
		return log.SeverityDebug
	default:
		return log.SeverityInfo
	}
}
