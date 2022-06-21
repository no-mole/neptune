package formatter

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/no-mole/neptune/logger/entry"
)

func init() {
	Registry("string", func() Formatter {
		return &FmtFormatter{}
	})
}

type FmtFormatter struct{}

func (f FmtFormatter) Format(entry entry.Entry) []byte {
	fields := entry.GetFields()
	keys := make([]string, len(fields))
	i := 0
	for k := range fields {
		keys[i] = k
		i++
	}
	buf := bytes.NewBufferString("")
	sort.Strings(keys)
	for _, k := range keys {
		buf.WriteString("|")
		buf.WriteString(k)
		buf.WriteString(fmt.Sprintf(":%+v", fields[k]))
	}
	buf.WriteString("|\n")
	return buf.Bytes()
}

var _ Formatter = &FmtFormatter{}
