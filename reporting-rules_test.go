package log

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestRepeatReportedNames(t *testing.T) {
	var a reportedNamesType
	c := qt.New(t)
	c.Assert(a.putReport([]string{"bunny"}), qt.IsTrue)
	c.Assert(a.putReport([]string{"bunny"}), qt.IsFalse)
	c.Assert(a.putReport([]string{"bunny", "foo", "foo"}), qt.IsTrue)
	c.Assert(a.putReport([]string{"bunny", "foo", "foo"}), qt.IsFalse)
	c.Assert(a.putReport([]string{"bunny", "foo"}), qt.IsTrue)
	c.Assert(a.putReport([]string{"bunny", "foo", "bar"}), qt.IsTrue)
	c.Assert(a.putReport([]string{"bunny", "foo", "bar"}), qt.IsFalse)
	c.Assert(a.putReport([]string{"bunny", "foo"}), qt.IsFalse)
	c.Assert(a.putReport([]string{"bunny"}), qt.IsFalse)
	c.Assert(a.putReport(nil), qt.IsTrue)
	c.Assert(a.putReport(nil), qt.IsFalse)
}
