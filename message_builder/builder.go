package message_builder

import (
	"encoding/json"
	"time"
)

type Builder interface {
	Type(t string) Builder
	Content(body string, color string, blob bool) Builder
	Data(key string, value interface{}) Builder
	Expire(expire int) Builder
	Build() []byte
}

func NewBuilder() Builder {
	return &ColorBuilder{
		Contents: []*Content{},
		DataMap:  map[string]interface{}{},
	}
}

type ColorBuilder struct {
	TypeStr    string                 `json:"type"`
	Contents   []*Content             `json:"content"`
	DataMap    map[string]interface{} `json:"data"`
	CreateTime string                 `json:"create_time"`
	ExpireTime int                    `json:"expire_time"`
}

type Content struct {
	Text  string `json:"text"`
	Color string `json:"color"`
	Blob  bool   `json:"blob"`
}

func (b *ColorBuilder) Type(t string) Builder {
	b.TypeStr = t
	return b
}

func (b *ColorBuilder) Content(body string, color string, blob bool) Builder {
	b.Contents = append(b.Contents, &Content{
		Text:  body,
		Color: color,
		Blob:  blob,
	})
	return b
}

func (b *ColorBuilder) Data(key string, value interface{}) Builder {
	b.DataMap[key] = value
	return b
}

func (b *ColorBuilder) Expire(expire int) Builder {
	b.ExpireTime = expire
	return b
}

func (b *ColorBuilder) Build() []byte {
	b.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	data, _ := json.Marshal(b)
	return data
}
