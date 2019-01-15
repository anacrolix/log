package log

import (
	"os"
	"testing"
)

func TestLogBadString(t *testing.T) {
	l := Logger{}
	// var buf bytes.Buffer
	l.SetHandler(&StreamHandler{W: os.Stderr, Fmt: LineFormatter})
	Str("\xef\x00\xaa\x1ctest\x00test").AddValue("\x00").Log(&l)
}
