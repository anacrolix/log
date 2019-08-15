package log

import (
	"bytes"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogBadString(t *testing.T) {
	l := Logger{}
	// var buf bytes.Buffer
	l.SetHandler(&StreamHandler{W: os.Stderr, Fmt: LineFormatter})
	Str("\xef\x00\xaa\x1ctest\x00test").AddValue("\x00").Log(&l)
}

type stringer struct {
	s string
}

func (me stringer) String() string {
	return strconv.Quote(me.s)
}

func TestValueStringNonLatin(t *testing.T) {
	s := stringer{"カワキヲアメク\n"}
	assert.Equal(t, `"カワキヲアメク\n"`, s.String())
	m := Str("").AddValue(&s)
	assert.Contains(t, string(LineFormatter(m)), `"カワキヲアメク\n"`)
	l := new(Logger)
	var buf bytes.Buffer
	l.SetHandler(&StreamHandler{&buf, LineFormatter})
	m.Log(l)
	assert.Contains(t, buf.String(), `"カワキヲアメク\n"`)
}
