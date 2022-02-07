package log

import (
	"os"
)

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
	Default.defaultLevel, err = levelFromString(os.Getenv("GO_LOG_DEFAULT_LEVEL"))
	if err != nil {
		panic(err)
	}
}
