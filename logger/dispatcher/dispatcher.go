package dispatcher

import (
	"errors"

	"github.com/no-mole/neptune/logger/entry"
	"github.com/no-mole/neptune/logger/formatter"
	"github.com/no-mole/neptune/logger/tagger"
)

type Dispatcher interface {
	tagger.Tagger
	formatter.Formatter
	Dispatch([]entry.Entry)
	SetFormatter(formatter.Formatter)
	SetTagger(tagger.Tagger)
}

type NewDispatcherFunc func(formatter formatter.Formatter, tagger tagger.Tagger, config *Config) Dispatcher

var dispatchers = map[string]NewDispatcherFunc{}

func Registry(name string, newFunc NewDispatcherFunc) {
	dispatchers[name] = newFunc
}

func NewDispatcher(name string, formatter formatter.Formatter, tagger tagger.Tagger, config *Config) (Dispatcher, error) {
	if fn, ok := dispatchers[name]; ok {
		return fn(formatter, tagger, config), nil
	}
	return nil, errors.New("no support dispatcher type: " + name)
}

var lineBreak = []byte("\n")
