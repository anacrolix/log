package log

import (
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
