package log_test

import (
	"io"
	"testing"

	"github.com/anacrolix/log"
)

// Mirrors usage seen for a particularly expensive logging callsite in anacrolix/dht.
func BenchmarkEmulateDhtServerReplyLogger(b *testing.B) {
	l := log.Default.FilterLevel(log.Info).WithValues(&struct{}{}).WithContextText("some dht prefix").WithDefaultLevel(log.Debug)
	makeMsg := func() log.Msg {
		return log.Fmsg(
			"reply to %q",
			"your mum", // dht.Addr caches its String method return value
		)
	}
	b.Run("Filtered", func(b *testing.B) {
		b.Run("Direct", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				makeMsg().Log(l)
			}
		})
		b.Run("Lazy", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				l.LazyLog(log.Debug, makeMsg)
			}
		})
	})
	b.Run("Unfiltered", func(b *testing.B) {
		h := log.DefaultHandler
		h.W = io.Discard
		l := l
		l.Handlers = []log.Handler{h}
		b.Run("Direct", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				makeMsg().LogLevel(log.Info, l)
			}
		})
		b.Run("Lazy", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				l.LazyLog(log.Info, makeMsg)
			}
		})
	})
}
