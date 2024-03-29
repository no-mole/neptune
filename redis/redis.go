package redis

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/extra/redisotel/v8"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/attribute"
	"gopkg.in/yaml.v3"
	"sync"
)

var (
	Client *client
)

func init() {
	Client = &client{}
}

type client struct {
	sync.Map
}

type Config struct {
	Host     string `json:"host" yaml:"host" validate:"required"`
	Database int    `json:"database" yaml:"database" validate:"required"`
	Password string `json:"password" yaml:"password" validate:"required"`
}

func InitWithConfig(redisName string, redisConf *Config) error {
	curClient := redis.NewClient(&redis.Options{
		Addr:     redisConf.Host,
		Password: redisConf.Password, // no password set
		DB:       redisConf.Database, // use default DB
	})
	// 发送一个ping命令,测试是否通
	_, err := curClient.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	curClient.AddHook(
		redisotel.NewTracingHook(
			redisotel.WithAttributes(
				attribute.String("redis_name", redisName),
				attribute.String("redis_host", redisConf.Host),
				attribute.Int("redis_db", redisConf.Database),
			),
		),
	)
	Client.StoreClient(redisName, curClient)
	return nil
}

// InitYaml 初始化redis yaml config string
func InitYaml(redisName string, confStr string) error {
	redisConf := &Config{}
	err := yaml.Unmarshal([]byte(confStr), redisConf)
	if err != nil {
		return err
	}
	return InitWithConfig(redisName, redisConf)
}

// Init 初始化redis json config string
func Init(redisName string, confStr string) error {
	redisConf := &Config{}
	err := json.Unmarshal([]byte(confStr), redisConf)
	if err != nil {
		return err
	}
	return InitWithConfig(redisName, redisConf)
}

func (c *client) StoreClient(key string, value *redis.Client) {
	c.Store(key, value)
}

func (c *client) GetClient(key string) (*redis.Client, bool) {
	if value, exist := c.Load(key); exist {
		if cc, ok := value.(*redis.Client); ok {
			return cc, true
		}
	}
	return nil, false
}
