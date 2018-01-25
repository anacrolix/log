package log

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var Default = new(Logger)

func init() {
	Default.SetHandler(&StreamHandler{
		W:   os.Stderr,
		Fmt: LineFormatter,
	})
}

type Logger struct {
	hs     map[Handler]struct{}
	values map[interface{}]struct{}
}

func (l *Logger) SetHandler(h Handler) {
	l.hs = map[Handler]struct{}{h: struct{}{}}
}

func (l *Logger) Emit(m Msg) {
	for h := range l.hs {
		h.Emit(m)
	}
}

type Msg struct {
	Values map[interface{}]struct{}
	text   string
}

func Fmsg(format string, a ...interface{}) Msg {
	return Msg{
		text: fmt.Sprintf(format, a...),
	}
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
	var ss []string
	for _, v := range msg.Values {
		ss = append(ss, fmt.Sprint(v))
	}
	return []byte(strings.Join(ss, " "))
}

func (me *StreamHandler) Emit(msg Msg) {
	me.W.Write(me.Fmt(msg))
}

func Printf(format string, a ...interface{}) {
	Default.Emit(Fmsg(format, a...))
}

func Print(v ...interface{}) {
	Default.Emit(Msg{text: fmt.Sprint(v...)})
}
