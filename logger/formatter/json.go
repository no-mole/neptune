package formatter

import (
	"fmt"

	"github.com/no-mole/neptune/json"
	"github.com/no-mole/neptune/logger/entry"
)

func init() {
	Registry("json", func() Formatter {
		return &JsonFormatter{}
	})
}

type JsonFormatter struct{}

func (j *JsonFormatter) Format(e entry.Entry) []byte {
	data, err := json.Marshal(e.GetFields())
	if err != nil {
		data, _ = json.Marshal(map[string]interface{}{"errorMsg": err.Error(), "input": fmt.Sprintf("%+v", e.GetFields())})
	}
	return data
}

var _ Formatter = &JsonFormatter{}
