package log

import (
	"fmt"
)

// Logger is a helper wrapping LoggerImpl.
type Logger struct {
	nonZero      bool
	names        []string
	values       []interface{}
	defaultLevel Level
	filterLevel  Level
	msgMaps      []func(Msg) Msg
	Handlers     []Handler
}

// Returns a logger that adds the given values to logged messages.
func (l Logger) WithValues(v ...interface{}) Logger {
	l.values = append(l.values, v...)
	return l
}

// Returns a logger that for a given message propagates the result of `f` instead.
func (l Logger) WithMap(f func(m Msg) Msg) Logger {
	l.msgMaps = append(l.msgMaps, f)
	return l
}

func (l Logger) WithText(f func(Msg) string) Logger {
	l.msgMaps = append(l.msgMaps, func(msg Msg) Msg {
		return msg.WithText(f)
	})
	return l
}

// Helper for compatibility with "log".Logger.
func (l Logger) Printf(format string, a ...interface{}) {
	l.LazyLog(l.defaultLevel, func() Msg {
		return Fmsg(format, a...).Skip(1)
	})
}

func (l Logger) Log(m Msg) {
	l.LogLevel(l.defaultLevel, m.Skip(1))
}

func (l Logger) LogLevel(level Level, m Msg) {
	l.LazyLog(level, func() Msg {
		return m.Skip(1)
	})
}

// Helper for compatibility with "log".Logger.
func (l Logger) Print(v ...interface{}) {
	l.LazyLog(l.defaultLevel, func() Msg {
		return Str(fmt.Sprint(v...)).Skip(1)
	})
}

func (l Logger) WithDefaultLevel(level Level) Logger {
	l.defaultLevel = level
	return l
}

func (l Logger) FilterLevel(minLevel Level) Logger {
	if _, ok := levelFromRules(l.names); !ok {
		l.filterLevel = minLevel
	}
	return l
}

func (l Logger) WithContextValue(v interface{}) Logger {
	return l.WithText(func(m Msg) string {
		return fmt.Sprintf("%v: %v", v, m)
	})
}

func (l Logger) WithContextText(s string) Logger {
	return l.WithText(func(m Msg) string {
		return s + ": " + m.Text()
	})
}

func (l Logger) IsZero() bool {
	return !l.nonZero
}

func (l Logger) SkipCallers(skip int) Logger {
	return l.WithMap(func(m Msg) Msg {
		return m.Skip(skip)
	})
}

func (l Logger) IsEnabledFor(level Level) bool {
	return !level.LessThan(l.filterLevel)
}

func (l Logger) LazyLog(level Level, f func() Msg) {
	l.lazyLog(level, 1, f)
}

func (l Logger) LazyLogDefaultLevel(f func() Msg) {
	l.lazyLog(l.defaultLevel, 1, f)
}

func (l Logger) lazyLog(level Level, skip int, f func() Msg) {
	if l.IsEnabledFor(level) {
		l.handle(level, f().Skip(skip+1))
	}
}

func (l Logger) handle(level Level, m Msg) {
	r := Record{
		Msg:   m.Skip(1),
		Level: level,
		Names: l.names,
	}
	if !l.nonZero {
		panic(fmt.Sprintf("Logger uninitialized. names=%q", l.names))
	}
	for _, h := range l.Handlers {
		h.Handle(r)
	}
}

func (l Logger) WithNames(names ...string) Logger {
	l.names = append(l.names, names...)
	return l.withFilterLevelFromRules()
}

func (l Logger) Levelf(level Level, format string, a ...interface{}) {
	l.LazyLog(level, func() Msg {
		return Fmsg(format, a...).Skip(1)
	})
}

func (l Logger) Println(a ...interface{}) {
	l.LazyLogDefaultLevel(func() Msg {
		return Str(fmt.Sprintln(a...)).Skip(1)
	})
}

func (l Logger) withFilterLevelFromRules() Logger {
	level, ok := levelFromRules(l.names)
	if ok {
		l.filterLevel = level
	}
	return l
}
