package log

import (
	"os"
)

var rules = []Rule{
	//func(names []string) (level Level, matched bool) {
	//	//log.Print(names)
	//	return Info, true
	//},
	//ContainsAllNames([]string{"reader"}, Debug),
}

type Rule func(names []string) (level Level, matched bool)

func stringSliceContains(s string, ss []string) bool {
	for _, sss := range ss {
		if s == sss {
			return true
		}
	}
	return false
}

func ContainsAllNames(all []string, level Level) Rule {
	return func(names []string) (_ Level, matched bool) {
		for _, s := range all {
			//log.Println(s, all, names)
			if !stringSliceContains(s, names) {
				return
			}
		}
		return level, true
	}
}

func parseEnvRules() (rules []Rule, err error) {
	rulesStr := os.Getenv("GO_LOG")
	if rulesStr == "" {
		return
	}
	var level Level
	err = level.UnmarshalText([]byte(rulesStr))
	if err != nil {
		return nil, err
	}
	return []Rule{
		func(names []string) (_ Level, matched bool) {
			return level, true
		},
	}, nil
}

func init() {
	var err error
	rules, err = parseEnvRules()
	if err != nil {
		panic(err)
	}
	Default = Logger{
		nonZero:     true,
		filterLevel: Error,
		Handlers:    []Handler{DefaultHandler},
	}.withFilterLevelFromRules()
}

func levelFromRules(names []string) (_ Level, ok bool) {
	// Later rules take precedence
	for i := len(rules) - 1; i >= 0; i-- {
		r := rules[i]
		level, ok := r(names)
		if ok {
			return level, true
		}
	}
	return
}
