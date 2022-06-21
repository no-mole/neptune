package bar

import (
	"time"

	"github.com/no-mole/neptune/protos/bar"
)

func (m *Model) Log(in, out string) (*bar.Bar, error) {
	item := &bar.Bar{
		In:         in,
		Out:        out,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	tx := m.db.Create(item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return item, nil
}
