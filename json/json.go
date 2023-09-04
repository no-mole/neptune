package json

import (
	jsoniter "github.com/json-iterator/go"
)

var Instance = jsoniter.ConfigCompatibleWithStandardLibrary

func Marshal(v interface{}) ([]byte, error) {
	return Instance.Marshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	return Instance.Unmarshal(data, v)
}

var NoTagInstance = jsoniter.Config{
	EscapeHTML:                    false,
	MarshalFloatWith6Digits:       true, // will lose precession
	ObjectFieldMustBeSimpleString: true, // do not unescape object field
	TagKey:                        "noJsonTag",
}.Froze()

func MarshalNoTag(v interface{}) ([]byte, error) {
	return NoTagInstance.Marshal(v)
}

func UnmarshalNoTag(data []byte, v interface{}) error {
	return NoTagInstance.Unmarshal(data, v)
}
