package dispatcher

import (
	"encoding/json"
	"testing"

	"github.com/no-mole/neptune/logger/level"

	"github.com/no-mole/neptune/logger/tagger"

	"github.com/no-mole/neptune/logger/entry"
	"github.com/no-mole/neptune/logger/formatter"
)

func TestHelper_SetFormatter(t *testing.T) {
	f, err := formatter.NewFormatter("json")
	if err != nil {
		t.Fatal(err)
	}
	e := entry.NewMapEntry(10, "app", level.LevelDebug)
	e.WithField("who", "Mark")

	h := NewHelper(f, nil)
	data := h.Format(e)
	expected, _ := json.Marshal(e.GetFields())
	if string(data) != string(expected) {
		t.Errorf("expected:%s,got:%s", string(expected), string(data))
	}
}

func TestHelper_SetTagger(t *testing.T) {
	tag := tagger.NewMatcher([]string{"app", "app1"})
	h := NewHelper(nil, tag)
	if !h.Match("app") {
		t.Errorf("match failed")
	}
	if !h.Match("app1") {
		t.Errorf("match failed")
	}
	if h.Match("app2") {
		t.Errorf("match failed")
	}
}
