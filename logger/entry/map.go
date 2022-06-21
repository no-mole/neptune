package entry

import "github.com/no-mole/neptune/logger/level"

func NewMapEntry(size int, tag string, lv level.Level) Entry {
	e := &MapEntry{
		tag:    tag,
		lv:     lv,
		fields: make(map[string]interface{}, size+2),
	}
	e.WithField("tag", tag)
	e.WithField("level", lv.String())
	return e
}

type MapEntry struct {
	tag    string
	lv     level.Level
	fields map[string]interface{}
}

func (e *MapEntry) GetLevel() level.Level {
	return e.GetLevel()
}

func (e *MapEntry) GetTag() string {
	return e.tag
}

func (e *MapEntry) WithField(key string, value interface{}) {
	e.fields[key] = value
}

func (e *MapEntry) GetFields() map[string]interface{} {
	return e.fields
}

func (e *MapEntry) WithFields(fields ...Field) {
	for _, fn := range fields {
		fn(e)
	}
}

var _ Entry = &MapEntry{}
