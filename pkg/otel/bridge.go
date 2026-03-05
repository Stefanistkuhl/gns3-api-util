package otel

import (
	"context"
	"log/slog"

	"go.uber.org/zap/zapcore"
)

type SlogCore struct {
	Handler slog.Handler
	Level   zapcore.LevelEnabler
}

func (c *SlogCore) Enabled(l zapcore.Level) bool         { return c.Level.Enabled(l) }
func (c *SlogCore) With(fs []zapcore.Field) zapcore.Core { return c }
func (c *SlogCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(e.Level) {
		return ce.AddCore(e, c)
	}
	return ce
}
func (c *SlogCore) Sync() error {
	return nil
}

func (c *SlogCore) Write(e zapcore.Entry, fs []zapcore.Field) error {
	slLevel := slog.LevelInfo
	switch e.Level {
	case zapcore.DebugLevel:
		slLevel = slog.LevelDebug
	case zapcore.WarnLevel:
		slLevel = slog.LevelWarn
	case zapcore.ErrorLevel, zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		slLevel = slog.LevelError
	}

	r := slog.NewRecord(e.Time, slLevel, e.Message, 0)
	for _, f := range fs {
		r.AddAttrs(slog.Any(f.Key, f.Interface))
	}
	return c.Handler.Handle(context.Background(), r)
}
