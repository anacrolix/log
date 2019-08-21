package log

import (
	"fmt"
	"io"
	"time"
)

type StreamLogger struct {
	W   io.Writer
	Fmt ByteFormatter
}

func (me StreamLogger) Log(msg Msg) {
	me.W.Write(me.Fmt(msg.Skip(1)))
}

type ByteFormatter func(Msg) []byte

func LineFormatter(msg Msg) []byte {
	var pc [1]uintptr
	msg.Callers(1, pc[:])
	ret := []byte(fmt.Sprintf(
		"%s %s: %s",
		time.Now().Format("2006-01-02 15:04:05"),
		humanPc(pc[0]),
		msg.Text(),
	))
	if ret[len(ret)-1] != '\n' {
		ret = append(ret, '\n')
	}
	return ret
}
