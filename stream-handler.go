package log

import (
	"fmt"
	"io"
	"path/filepath"
)

type StreamHandler struct {
	W   io.Writer
	Fmt ByteFormatter
}

func (me StreamHandler) Handle(r Record) {
	r.Msg = r.Skip(1)
	me.W.Write(me.Fmt(r))
}

type ByteFormatter func(Record) []byte

func getMsgPcName(msg Msg) string {
	var pc [1]uintptr
	msg.Callers(1, pc[:])
	return pcName(pc[0])
}

func LineFormatter(msg Record) []byte {
	names := msg.Names
	ret := []byte(fmt.Sprintf(
		"%s %s %s: %s",
		DefaultTimeFormatter(),
		msg.Level.LogString(),
		names,
		msg.Text(),
	))
	if ret[len(ret)-1] != '\n' {
		ret = append(ret, '\n')
	}
	return ret
}

func pcName(pc uintptr) string {
	if pc == 0 {
		panic(pc)
	}
	loc := locFromPc(pc)
	return fmt.Sprintf("%v:%v:%v", loc.Package, filepath.Base(loc.File), loc.Line)
}

func pcNames(pc uintptr, names []string) []string {
	return append(names, pcName(pc))
}
