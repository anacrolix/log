package log

func Printf(format string, a ...interface{}) {
	Default.Log(Fmsg(format, a...).Skip(1))
}
