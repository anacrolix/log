package log

import (
	"encoding"
	"fmt"
	"strconv"
	"strings"
)

type Level struct {
	rank int
}

var levelKey = new(struct{})

var (
	Disabled = Level{-1} // This would be no filtering?
	NotSet   = Level{0}
	Debug    = Level{1}
	Info     = Level{2}
	Warning  = Level{3}
	Error    = Level{4}
	Critical = Level{5}
	// Will this get special treatment? Not yet. Also disabled due to conflict with std log.Fatal.
	//Fatal = Level{6, "FATAL"}
)

func (l Level) isNotSet() bool {
	return l.rank == 0
}

func (l Level) LogString() string {
	switch l.rank {
	case NotSet.rank:
		return "NIL"
	case Debug.rank:
		return "DBG"
	case Info.rank:
		return "INF"
	case Warning.rank:
		return "WRN"
	case Error.rank:
		return "ERR"
	case Critical.rank:
		return "CRT"
	//case Fatal.rank:
	//	return "fatal"
	default:
		return strconv.FormatInt(int64(l.rank), 10)
	}
}

func (l Level) LessThan(r Level) bool {
	if l.rank == NotSet.rank {
		return false
	}
	return l.rank < r.rank
}

var _ encoding.TextUnmarshaler = (*Level)(nil)

func (l *Level) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "nil", "notset", "unset", "all":
		*l = NotSet
	case "dbg", "debug":
		*l = Debug
	case "inf", "info":
		*l = Info
	case "wrn", "warning", "warn":
		*l = Warning
	case "err", "error":
		*l = Error
	case "crt", "critical", "crit":
		*l = Critical
	//case "FATAL":
	//	*l = Fatal
	default:
		return fmt.Errorf("unknown log level: %q", text)
	}
	return nil
}
