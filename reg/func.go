package reg

import (
	"fmt"
	"runtime"
	"strings"
)

// getFnNameFile returns the first file+function in the call stack that is not avo.
func getFnNameFile(depth int) string {
	caller := [100]uintptr{0}
	runtime.Callers(1+depth, caller[:])
	frames := runtime.CallersFrames(caller[:])
	for {
		f, more := frames.Next()
		file := f.File
		line := f.Line
		fnName := f.Func.Name()

		if !strings.Contains(file, "github.com/mmcloughlin/avo") {
			return fmt.Sprintf("%s:%d %s()", file, line, fnName)
		}
		if !more {
			break
		}
	}
	return ""
}
