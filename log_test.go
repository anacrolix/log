package log

import (
	"errors"
	"runtime"
	"strconv"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/stretchr/testify/assert"
)

func TestLogBadString(t *testing.T) {
	Str("\xef\x00\xaa\x1ctest\x00test").AddValue("\x00").Log(Default)
}

type stringer struct {
	s string
}

func (me stringer) String() string {
	return strconv.Quote(me.s)
}

func TestValueStringNonLatin(t *testing.T) {
	const (
		u = "カワキヲアメク\n"
		q = `"カワキヲアメク\n"`
	)
	s := stringer{u}
	assert.Equal(t, q, s.String())
	m := Str("").AddValue(q)
	assert.True(t, m.HasValue(q))
}

func BenchmarkPcNames(b *testing.B) {
	b.ReportAllocs()
	var names []string
	for i := 0; i < b.N; i++ {
		var pc [1]uintptr
		runtime.Callers(1, pc[:])
		names = pcNames(pc[0], names[:0])
		//b.Log(names[0], names[1])
		//panic("hi")
	}
}

type chanHandler struct {
	r chan<- Record
}

func (c chanHandler) Handle(r Record) {
	c.r <- r
}

func TestErrorLevelHandling(t *testing.T) {
	c := qt.New(t)
	l := NewLogger("test").FilterLevel(NotSet)
	rs := make(chan Record)
	// We could use SetHandlers here, but it's nice to see the output in verbose testing mode.
	l.Handlers = append(l.Handlers, chanHandler{rs})
	checkRecord := func(expectedLevel Level) {
		r := <-rs
		c.Check(r.Level, qt.Equals, expectedLevel, qt.Commentf("message received: %v", r.Msg))
	}
	testLogging := func(expectedLevel Level, logAction func()) {
		go logAction()
		checkRecord(expectedLevel)
	}
	testLogging(l.defaultLevel, func() { l.Printf("should have default level") })
	testLogging(Info, func() { l.WithDefaultLevel(Info).Printf("should be info") })
	testLogging(l.defaultLevel, func() { l.Levelf(NotSet, "should be starting level") })
	testLogging(Info, func() { l.Levelf(Info, "should be info") })
	err := errors.New("oh no something broke")
	testLogging(l.defaultLevel, func() { l.Levelf(ErrorLevel(err), "error without level: %v", err) })
	testLogging(Warning, func() { l.Levelf(ErrorLevel(WithLevel(Warning, err)), "error with level: %v", err) })
}
