package dispatcher

import (
	"os"

	"github.com/no-mole/neptune/logger/tagger"

	"github.com/no-mole/neptune/logger/entry"
	"github.com/no-mole/neptune/logger/formatter"
)

func init() {
	Registry("stdout", func(formatter formatter.Formatter, tagger tagger.Tagger, config *Config) Dispatcher {
		return &StdoutDispatcher{Helper: NewHelper(formatter, tagger)}
	})
}

type StdoutDispatcher struct {
	Helper
}

func (s *StdoutDispatcher) Dispatch(entries []entry.Entry) {
	for _, e := range entries {
		if !s.Match(e.GetTag()) {
			continue
		}
		_, _ = os.Stdout.Write(s.Format(e))
		_, _ = os.Stdout.Write(lineBreak)
	}
}

var _ Dispatcher = &StdoutDispatcher{}
