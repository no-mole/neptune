package dispatcher

import (
	"github.com/no-mole/neptune/logger/formatter"
	"github.com/no-mole/neptune/logger/tagger"
)

func NewHelper(formatter formatter.Formatter, tagger tagger.Tagger) Helper {
	return Helper{Formatter: formatter, Tagger: tagger}
}

type Helper struct {
	tagger.Tagger
	formatter.Formatter
}

func (h *Helper) SetFormatter(formatter formatter.Formatter) {
	h.Formatter = formatter
}

func (h *Helper) SetTagger(tagger tagger.Tagger) {
	h.Tagger = tagger
}
