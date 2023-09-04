package config

import (
	"context"
	"fmt"
	"sync"
)

type RegistryImplementationFunc func(ctx context.Context) Client

var configCenterImplementation = map[string]RegistryImplementationFunc{}

func RegistryImplementation(typeName string, fn RegistryImplementationFunc) {
	configCenterImplementation[typeName] = fn
}

type Config struct {
	Type      string `yaml:"type" json:"type"`
	Endpoints string `yaml:"endpoints" json:"endpoints"`
	Namespace string `yaml:"namespace" json:"namespace"`

	Username string `yaml:"username" json:"userName"`
	Password string `yaml:"password" json:"password"`

	Settings map[string]string `yaml:"settings" json:"settings"`
}

type Client interface {
	Close() error
	Init(ctx context.Context, config *Config) error
	Set(ctx context.Context, key, value string) error
	Get(ctx context.Context, key string) (*Item, error)
	Watch(ctx context.Context, item *Item, callback func(item *Item)) error
	Exist(ctx context.Context, key string) (bool, error)
}

// GetClientImplementation get client implementation by typeName
func GetClientImplementation(ctx context.Context, typeName string) (Client, error) {
	if fn, ok := configCenterImplementation[typeName]; ok {
		return fn(ctx), nil
	}
	return nil, fmt.Errorf("no implementation config client for %s", typeName)
}

func NewItem(namespace, key, value string) *Item {
	return &Item{
		Namespace: namespace,
		Key:       key,
		value:     value,
		RWMutex:   &sync.RWMutex{},
	}
}

type Item struct {
	Namespace     string `json:"namespace"`
	Key           string `json:"key"`
	value         string
	*sync.RWMutex `json:"-"`
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

var defaultClient Client

func Init(ctx context.Context, config *Config) error {
	impl, err := GetClientImplementation(ctx, config.Type)
	if err != nil {
		return err
	}
	defaultClient = impl
	return impl.Init(ctx, config)
}
func Set(ctx context.Context, key, value string) error {
	return defaultClient.Set(ctx, key, value)
}
func Get(ctx context.Context, key string) (*Item, error) {
	return defaultClient.Get(ctx, key)
}
func Watch(ctx context.Context, item *Item, callback func(item *Item)) error {
	return defaultClient.Watch(ctx, item, callback)
}
func Exist(ctx context.Context, key string) (bool, error) {
	return defaultClient.Exist(ctx, key)
}

func Close() error {
	return defaultClient.Close()
}
