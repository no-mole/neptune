package env

import (
	"os"
	"strings"
)

type SetFunc func(value string)

type Item struct {
	Key    string
	Setter SetFunc
}

func Load(items ...*Item) {
	for _, item := range items {
		item.Setter(strings.TrimSpace(os.Getenv(item.Key)))
	}
}
