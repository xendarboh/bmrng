package utils

import (
	"fmt"
	"log"
	"regexp"
	"runtime"
)

var debugLogEnabled = false

func SetDebugLogEnabled(enabled bool) {
	debugLogEnabled = enabled
}

func DebugLog(format string, args ...interface{}) {
	if !debugLogEnabled {
		return
	}

	// get the calling function's name, filename, and line number
	pc, filename, line, _ := runtime.Caller(1)
	funcname := runtime.FuncForPC(pc).Name()

	// remove project root to shorten file path
	re := regexp.MustCompile("(.*)/trellis/")
	path := re.ReplaceAllString(filename, "")

	s1 := fmt.Sprintf("🔷 %s:%s:%d", path, funcname, line)
	s2 := fmt.Sprintf(format, args...)
	log.Printf("%s %s", s1, s2)
}
