package entry

import "github.com/no-mole/neptune/logger/level"

type Field func(Entry)

type Entry interface {
	WithField(key string, value interface{})
	WithFields(fields ...Field)
	GetFields() map[string]interface{}
	GetTag() string
	GetLevel() level.Level
}
