package log

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/anacrolix/missinggo/iter"
)

type Msg struct {
	MsgImpl
}

type MsgImpl interface {
	Text() string
	Callers(skip int, pc []uintptr) int
	Values(callback iter.Callback)
}

// maybe implement finalizer to ensure msgs are sunk
type rootMsg struct {
	text string
}

func (m rootMsg) Text() string {
	return m.text
}

func (m rootMsg) Callers(skip int, pc []uintptr) int {
	return runtime.Callers(skip+2, pc)
}

func (m rootMsg) Values(iter.Callback) {}

func newMsgWithCallers(text string, skip int) Msg {
	return Msg{rootMsg{text}}
}

func Fmsg(format string, a ...interface{}) Msg {
	return newMsgWithCallers(fmt.Sprintf(format, a...), 1)
}

func Str(s string) (m Msg) {
	return newMsgWithCallers(s, 1)
}

type msgSkipCaller struct {
	MsgImpl
	skip int
}

func (me msgSkipCaller) Callers(skip int, pc []uintptr) int {
	return me.MsgImpl.Callers(skip+1+me.skip, pc)
}

func (m Msg) Skip(skip int) Msg {
	return Msg{msgSkipCaller{m.MsgImpl, skip}}
}

type item struct {
	key, value interface{}
}

// rename sink
func (msg Msg) Log(l Logger) Msg {
	l.Log(msg.Skip(1))
	return msg
}

type msgWithValues struct {
	MsgImpl
	values []interface{}
}

func (me msgWithValues) Values(cb iter.Callback) {
	for _, v := range me.values {
		if !cb(v) {
			return
		}
	}
	me.MsgImpl.Values(cb)
}

func (me Msg) WithValues(v ...interface{}) Msg {
	return Msg{msgWithValues{me.MsgImpl, v}}
}

func (me Msg) AddValues(v ...interface{}) Msg {
	return me.WithValues(v...)
}

func (me Msg) With(key, value interface{}) Msg {
	return me.WithValues(item{key, value})
}

func (me Msg) Add(key, value interface{}) Msg {
	return me.With(key, value)
}

func (me Msg) HasValue(v interface{}) (has bool) {
	me.Values(func(i interface{}) bool {
		if i == v {
			has = true
		}
		return !has
	})
	return
}

func (me Msg) AddValue(v interface{}) Msg {
	return me.AddValues(v)
}

func humanPc(pc uintptr) string {
	if pc == 0 {
		panic(pc)
	}
	f, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	return fmt.Sprintf("%s:%d", filepath.Base(f.File), f.Line)
}
