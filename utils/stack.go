package utils

import (
	"bytes"
	"fmt"
	"runtime"
)

func GetStack(skip int, depth int) string {
	buf := bytes.NewBufferString("")
	pcs := make([]uintptr, depth)
	n := runtime.Callers(skip, pcs)
	for _, pc := range pcs[:n] {
		f := runtime.FuncForPC(pc)
		file, line := f.FileLine(pc)
		buf.WriteString(fmt.Sprintf("%s %d\n", file, line))
	}
	return buf.String()
}
