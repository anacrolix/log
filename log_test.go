package log

import (
	"runtime"
	"strconv"
	"testing"

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
