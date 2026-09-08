package log

import (
	"testing"

	"github.com/go-quicktest/qt"
)

func TestRepeatReportedNames(t *testing.T) {
	var a reportedNamesType
	qt.Assert(t, qt.IsTrue(a.putReport([]string{"bunny"})))
	qt.Assert(t, qt.IsFalse(a.putReport([]string{"bunny"})))
	qt.Assert(t, qt.IsTrue(a.putReport([]string{"bunny", "foo", "foo"})))
	qt.Assert(t, qt.IsFalse(a.putReport([]string{"bunny", "foo", "foo"})))
	qt.Assert(t, qt.IsTrue(a.putReport([]string{"bunny", "foo"})))
	qt.Assert(t, qt.IsTrue(a.putReport([]string{"bunny", "foo", "bar"})))
	qt.Assert(t, qt.IsFalse(a.putReport([]string{"bunny", "foo", "bar"})))
	qt.Assert(t, qt.IsFalse(a.putReport([]string{"bunny", "foo"})))
	qt.Assert(t, qt.IsFalse(a.putReport([]string{"bunny"})))
	qt.Assert(t, qt.IsTrue(a.putReport(nil)))
	qt.Assert(t, qt.IsFalse(a.putReport(nil)))
}
