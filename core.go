package core

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Sentry struct {
	zapcore.Core
	fields []zapcore.Field
}

func WithSentry(options sentry.ClientOptions) zap.Option {
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		err := sentry.Init(options)
		if err != nil {
			panic(fmt.Errorf("failed to create sentry client: %w", err))
		}
		s := Sentry{core, nil}
		return zapcore.NewTee(core, s)
	})
}

func (s Sentry) Enabled(level zapcore.Level) bool {
	return s.Core.Enabled(level)
}

func (s Sentry) With(fields []zapcore.Field) zapcore.Core {
	return Sentry{s.Core.With(fields), fields}
}

func (s Sentry) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if s.Enabled(entry.Level) {
		return ce.AddCore(entry, s)
	}

	return ce
}

func (s Sentry) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	m := make(map[string]interface{}, len(fields))

	s.fields = append(s.fields, fields...)

	enc := zapcore.NewMapObjectEncoder()

	for _, f := range s.fields {
		f.AddTo(enc)
	}

	// Merge the two maps.
	for k, v := range enc.Fields {
		m[k] = v
	}

	hub := sentry.CurrentHub().Clone()

	event := sentry.NewEvent()
	event.Level = levelTransformer(entry.Level)
	event.Message = entry.Message
	event.Timestamp = entry.Time
	event.Tags["service"] = entry.LoggerName
	event.Extra = m

	hub.CaptureEvent(event)

	return nil
}

func levelTransformer(level zapcore.Level) sentry.Level {
	switch level {
	case zapcore.DebugLevel:
		return sentry.LevelDebug
	case zapcore.InfoLevel:
		return sentry.LevelInfo
	case zapcore.WarnLevel:
		return sentry.LevelWarning
	case zapcore.ErrorLevel:
		return sentry.LevelError
	case zapcore.DPanicLevel:
		return sentry.LevelFatal
	case zapcore.PanicLevel:
		return sentry.LevelFatal
	case zapcore.FatalLevel:
		return sentry.LevelFatal
	default:
		return sentry.LevelInfo
	}
}

func (s Sentry) Sync() error {
	const flushTime = 2 * time.Second

	sentry.Flush(flushTime)

	return nil
}
