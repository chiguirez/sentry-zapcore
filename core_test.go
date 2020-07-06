package core_test

import (
	"errors"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/chiguirez/sentry-zapcore"
)

func TestSentry(t *testing.T) {
	noop := &TransportNOOP{}
	t.Run("Given a Zap logger withSentry Option", func(t *testing.T) {
		sut := zap.New(NewNopCore()).WithOptions(core.WithSentry(sentry.ClientOptions{
			Transport: noop,
		}))
		t.Run("When something is logged", func(t *testing.T) {
			const msg = "my custom error"
			err := errors.New("custom internal error")
			now := time.Now()
			sut.With(zap.Time("timeStamp", now)).Error(msg, zap.Error(err))
			t.Run("Then its also logged on sentry", func(t *testing.T) {
				assert.Equal(t, noop.lastEvent.Message, msg)
				assert.Equal(t, noop.lastEvent.Extra["error"], err.Error())
				assert.Equal(t, noop.lastEvent.Extra["timeStamp"].(time.Time).Unix(), now.Unix())
			})
		})
		err := sut.Sync()
		require.NoError(t, err)
	})
}

type TransportNOOP struct {
	lastEvent *sentry.Event
}

func (t *TransportNOOP) Flush(_ time.Duration) bool       { return true }
func (t *TransportNOOP) Configure(_ sentry.ClientOptions) {}
func (t *TransportNOOP) SendEvent(e *sentry.Event)        { t.lastEvent = e }

type nopCore struct{}

func NewNopCore() zapcore.Core                                                        { return nopCore{} }
func (nopCore) Enabled(zapcore.Level) bool                                            { return true }
func (n nopCore) With([]zap.Field) zapcore.Core                                       { return n }
func (nopCore) Check(_ zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry { return ce }
func (nopCore) Write(zapcore.Entry, []zap.Field) error                                { return nil }
func (nopCore) Sync() error                                                           { return nil }
