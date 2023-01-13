package log

import (
	"log"
	"sync"

	"github.com/anacrolix/generics"
)

type nameToAny struct {
	emptyCase bool
	children  map[string]*nameToAny
}

var reportedNames reportedNamesType

type reportedNamesType struct {
	mu   sync.Mutex
	base nameToAny
}

func putReportInner(toAny *nameToAny, names []string) bool {
	if len(names) == 0 {
		if toAny.emptyCase {
			return false
		}
		toAny.emptyCase = true
		return true
	}
	generics.MakeMapIfNil(&toAny.children)
	child, ok := toAny.children[names[0]]
	if !ok {
		child = new(nameToAny)
		toAny.children[names[0]] = child
	}
	return putReportInner(child, names[1:])
}

// Prevent duplicate logs about the same series of names.
func (me *reportedNamesType) putReport(names []string) bool {
	me.mu.Lock()
	defer me.mu.Unlock()
	return putReportInner(&me.base, names)
}

func reportLevelFromRules(level Level, ok bool, names []string) {
	if !reportedNames.putReport(names) {
		return
	}
	if !ok {
		log.Printf("no rule matched for %q", names)
		return
	}
	log.Printf("got level %v for %q", level.LogString(), names)
}
