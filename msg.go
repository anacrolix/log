package log

import "fmt"

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

func Str(s string) Msg {
	return Msg{
		text: s,
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

func (m Msg) AddValues(values ...interface{}) Msg {
	for _, v := range values {
		m = m.AddValue(v)
	}
	return m
}
