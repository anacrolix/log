package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
)

var Default = new(Logger)

func init() {
	Default.SetHandler(&StreamHandler{
		W:   os.Stderr,
		Fmt: LineFormatter,
	})
}

type Logger struct {
	hs      map[Handler]struct{}
	values  map[interface{}]struct{}
	filters map[*Filter]struct{}
}

func (l *Logger) SetHandler(h Handler) {
	l.hs = map[Handler]struct{}{h: struct{}{}}
}

func (l *Logger) Clone() *Logger {
	ret := &Logger{
		hs:      make(map[Handler]struct{}),
		values:  make(map[interface{}]struct{}),
		filters: make(map[*Filter]struct{}),
	}
	for h, v := range l.hs {
		ret.hs[h] = v
	}
	for v, v_ := range l.values {
		ret.values[v] = v_
	}
	for f := range l.filters {
		ret.filters[f] = struct{}{}
	}
	return ret
}

func (l *Logger) AddValue(v interface{}) *Logger {
	l.values[v] = struct{}{}
	return l
}

func (l *Logger) Handle(m Msg) {
	for v := range l.values {
		m.AddValue(v)
	}
	for f := range l.filters {
		if !f.ff(&m) {
			return
		}
	}
	for h := range l.hs {
		h.Emit(m)
	}
}

func (l *Logger) AddFilter(f *Filter) *Logger {
	if l.filters == nil {
		l.filters = make(map[*Filter]struct{})
	}
	l.filters[f] = struct{}{}
	return l
}

type Msg struct {
	fields map[string][]interface{}
	values map[interface{}]struct{}
	text   string
}

func Fmsg(format string, a ...interface{}) Msg {
	return Msg{
		text: fmt.Sprintf(format, a...),
	}
}

func (msg Msg) Add(key string, value interface{}) Msg {
	if msg.fields == nil {
		msg.fields = make(map[string][]interface{})
	}
	msg.fields[key] = append(msg.fields[key], value)
	return msg
}

func (msg Msg) Log(l *Logger) Msg {
	l.Handle(msg)
	return msg
}

func (m Msg) Values() map[interface{}]struct{} {
	return m.values
}

func (m Msg) AddValue(value interface{}) Msg {
	if m.values == nil {
		m.values = make(map[interface{}]struct{})
	}
	m.values[value] = struct{}{}
	return m
}

type Handler interface {
	Emit(Msg)
}

type ByteFormatter func(Msg) []byte

type StreamHandler struct {
	W   io.Writer
	Fmt ByteFormatter
}

func LineFormatter(msg Msg) []byte {
	ret := []byte(fmt.Sprintf("%s, %v, %v\n", msg.text, msg.values, msg.fields))
	if ret[len(ret)-1] != '\n' {
		ret = append(ret, '\n')
	}
	return ret
}

func (me *StreamHandler) Emit(msg Msg) {
	me.W.Write(me.Fmt(msg))
}

func Printf(format string, a ...interface{}) {
	Default.Handle(Fmsg(format, a...))
}

func Print(v ...interface{}) {
	Default.Handle(Msg{text: fmt.Sprint(v...)})
}

func Call() Msg {
	var pc [1]uintptr
	n := runtime.Callers(4, pc[:])
	fs := runtime.CallersFrames(pc[:n])
	f, _ := fs.Next()
	return Fmsg("called %q", f.Function)
}
