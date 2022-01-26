package log

import (
	"strconv"
)

type Level struct {
	rank int
}

var levelKey = new(struct{})

var (
	NotSet   = Level{0}
	Debug    = Level{1}
	Info     = Level{2}
	Warning  = Level{3}
	Error    = Level{4}
	Critical = Level{5}
	// Will this get special treatment? Not yet. Also disabled due to conflict with std log.Fatal.
	//Fatal = Level{6, "FATAL"}
)

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
