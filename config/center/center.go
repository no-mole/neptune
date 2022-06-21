package center

import (
	"context"
	"errors"
	"sync"
)

type Client interface {
	Close() error

	Init(opts ...Option) error

	Set(ctx context.Context, key, value string) error
	SetEX(ctx context.Context, key, value string, ttl int64) error
	SetExKeepAlive(ctx context.Context, key, value string, ttl int64) error

	Get(ctx context.Context, key string) (*Item, error)
	GetDefault(ctx context.Context, key string, defaultValue string) (*Item, error)
	GetWithPrefixKey(ctx context.Context, prefixKey string) (*Item, error)

	Exist(ctx context.Context, key string) (bool, error)

	Watch(ctx context.Context, item *Item, callback func(item *Item))
	WatchWithPrefix(ctx context.Context, item *Item, callback func(item *Item))
}

var configCenterImplementation = map[string]Client{}

func RegistryImplementation(typeName string, cli Client) {
	configCenterImplementation[typeName] = cli
}

//GetClient get client implementation by typeName
func GetClient(typeName string) (Client, error) {
	if cli, ok := configCenterImplementation[typeName]; ok {
		return cli, nil
	}
	return nil, errors.New("")
}

//Item
type Item struct {
	Act          int64
	Namespace    string `json:"namespace"`
	Key          string `json:"key"`
	value        string
	Kvs          []*KVs
	IsDefault    bool `json:"is_default"`
	sync.RWMutex `json:"-"`
}

type KVs struct {
	Key   string
	Value string
}

func (i *Item) SetValue(value string) {
	i.Lock()
	defer i.Unlock()
	i.value = value
}

func (i *Item) GetValue() string {
	i.RLock()
	defer i.RUnlock()
	return i.value
}
