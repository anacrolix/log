package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"time"
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

var timeFmt string

func init() {
	var ok bool
	timeFmt, ok = os.LookupEnv("GO_LOG_TIME_FMT")
	if !ok {
		timeFmt = "2006-01-02T15:04:05-0700"
	}
	if timeFmt != "" {
		timeFmt += " "
	}
}

func getMsgPcName(msg Msg) string {
	var pc [1]uintptr
	msg.Callers(1, pc[:])
	return pcName(pc[0])
}

func LineFormatter(msg Record) []byte {
	names := msg.Names
	ret := []byte(fmt.Sprintf(
		"%s%s %s: %s",
		time.Now().Format(timeFmt),
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
	funcName, file, line := func() (string, string, int) {
		if false {
			// This seems to result in one less allocation, but doesn't handle inlining?
			func_ := runtime.FuncForPC(pc)
			file, line := func_.FileLine(pc)
			return func_.Name(), file, line
		} else {
			f, _ := runtime.CallersFrames([]uintptr{pc}).Next()
			return f.Function, f.File, f.Line
		}
	}()
	_ = file
	return fmt.Sprintf("%s:%v", funcName, line)
}

func pcNames(pc uintptr, names []string) []string {
	return append(names, pcName(pc))
}
