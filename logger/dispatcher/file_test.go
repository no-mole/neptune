package dispatcher

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/no-mole/neptune/logger/level"

	"github.com/no-mole/neptune/logger/entry"

	"github.com/no-mole/neptune/logger/tagger"

	"github.com/no-mole/neptune/logger/formatter"
)

func TestFileDispatcher_Dispatch(t *testing.T) {
	f, err := formatter.NewFormatter("json")
	if err != nil {
		t.Fatal(err.Error())
	}

	file, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err.Error())
	}
	instance, err := NewDispatcher(
		"file",
		f,
		tagger.NewMatcher([]string{"*"}),
		&Config{Fields: map[string]string{
			"fileName": file.Name(),
		}},
	)
	if err != nil {
		t.Error(err.Error())
	}
	e1 := entry.NewMapEntry(10, "tag1", level.LevelDebug)
	e1.WithField("key1", "value1")
	instance.Dispatch([]entry.Entry{e1})
	time.Sleep(time.Second)
	body, err := os.ReadFile(file.Name())
	if err != nil {
		t.Error(err.Error())
	}
	want := &struct {
		Tag  string `json:"tag"`
		Key1 string `json:"key1"`
	}{}

	err = json.Unmarshal([]byte(strings.TrimSuffix(string(body), "\n")), want)
	if err != nil {
		t.Fatal(err.Error())
	}
	if want.Tag != "tag1" || want.Key1 != "value1" {
		t.Errorf("%+v", want)
	}
}
