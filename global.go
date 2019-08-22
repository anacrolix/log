package log

import (
	"fmt"
	"os"
)

var Default = Logger{StreamLogger{
	W:   os.Stderr,
	Fmt: LineFormatter,
}}

func Printf(format string, a ...interface{}) {
	Default.Log(Fmsg(format, a...).Skip(1))
}

// Prints the arguments to the Default Logger.
func Print(a ...interface{}) {
	// TODO: There's no "Print" equivalent constructor for a Msg, and I don't know what I'd call it.
	Str(fmt.Sprint(a...)).Skip(1).Log(Default)
}
