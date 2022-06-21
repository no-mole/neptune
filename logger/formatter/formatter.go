package formatter

import (
	"errors"

	"github.com/no-mole/neptune/logger/entry"
)

type Formatter interface {
	Format(entry.Entry) []byte
}

type NewFormatterFunc func() Formatter

var newFormatterFuncMapping = map[string]NewFormatterFunc{}

func Registry(name string, fn NewFormatterFunc) {
	newFormatterFuncMapping[name] = fn
}

func NewFormatter(name string) (Formatter, error) {
	if fn, ok := newFormatterFuncMapping[name]; ok {
		return fn(), nil
	}
	return nil, errors.New("no support formatter name:" + name)
}
