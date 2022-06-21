package json

import (
	jsoniter "github.com/json-iterator/go"
)

var JsonInstance = jsoniter.ConfigCompatibleWithStandardLibrary

func Marshal(v interface{}) ([]byte, error) {
	return JsonInstance.Marshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	return JsonInstance.Unmarshal(data, v)
}

var JsonNoTagInstance = jsoniter.Config{
	EscapeHTML:                    false,
	MarshalFloatWith6Digits:       true, // will lose precession
	ObjectFieldMustBeSimpleString: true, // do not unescape object field
	TagKey:                        "noJsonTag",
}.Froze()

func MarshalNoTag(v interface{}) ([]byte, error) {
	return JsonNoTagInstance.Marshal(v)
}

func UnmarshalNoTag(data []byte, v interface{}) error {
	return JsonNoTagInstance.Unmarshal(data, v)
}
