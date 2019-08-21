package log

import "os"

var Default = Logger{StreamLogger{
	W:   os.Stderr,
	Fmt: LineFormatter,
}}

func Printf(format string, a ...interface{}) {
	Default.Log(Fmsg(format, a...).Skip(1))
}
