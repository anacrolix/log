package log

import (
	"fmt"

	"github.com/anacrolix/missinggo/iter"
)

type Msg struct {
	MsgImpl
}

func newMsg(text string) Msg {
	return Msg{rootMsgImpl{text}}
}

func Fmsg(format string, a ...interface{}) Msg {
	return newMsg(fmt.Sprintf(format, a...))
}

var Fstr = Fmsg

func Str(s string) (m Msg) {
	return newMsg(s)
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
func (m Msg) Log(l Logger) Msg {
	l.Log(m.Skip(1))
	return m
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

// TODO: What ordering should be applied to the values here, per MsgImpl.Values. For now they're
// traversed in order of the slice.
func (m Msg) WithValues(v ...interface{}) Msg {
	return Msg{msgWithValues{m.MsgImpl, v}}
}

func (m Msg) AddValues(v ...interface{}) Msg {
	return m.WithValues(v...)
}

func (m Msg) With(key, value interface{}) Msg {
	return m.WithValues(item{key, value})
}

func (m Msg) Add(key, value interface{}) Msg {
	return m.With(key, value)
}

func (m Msg) HasValue(v interface{}) (has bool) {
	m.Values(func(i interface{}) bool {
		if i == v {
			has = true
		}
		return !has
	})
	return
}

func (m Msg) AddValue(v interface{}) Msg {
	return m.AddValues(v)
}
