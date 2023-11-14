package rabbitmq

import (
	"encoding/json"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/sync/singleflight"
	"sync"
	"time"
)

var (
	Client    *client
	configMap map[string]*Config
	g         singleflight.Group
)

func init() {
	Client = &client{}
	configMap = make(map[string]*Config)
	g = singleflight.Group{}
}

type client struct {
	sync.Map
}

type Config struct {
	Host     string `json:"host" yaml:"host" validate:"required"`
	Username string `json:"username" yaml:"username" validate:"required"`
	Password string `json:"password" yaml:"password" validate:"required"`
	Vhosts   string `json:"vhosts" yaml:"vhosts" validate:"required"`
}

func InitRabbitMq(rabbitMqName string, confStr string) error {
	mqConf := &Config{}
	err := json.Unmarshal([]byte(confStr), mqConf)
	if err != nil {
		return err
	}
	configMap[rabbitMqName] = mqConf
	conn := getRabbitMqConn(mqConf)
	Client.StoreClient(rabbitMqName, conn)
	return nil
}

func InitRabbitMqWithConfig(rabbitMqName string, conf *Config) error {
	configMap[rabbitMqName] = conf
	conn := getRabbitMqConn(conf)
	Client.StoreClient(rabbitMqName, conn)
	return nil
}

func getRabbitMqConn(mqConf *Config) *amqp.Connection {
	url := fmt.Sprintf("amqp://%s:%s@%s/", mqConf.Username, mqConf.Password, mqConf.Host)
	conn, err := amqp.DialConfig(url, amqp.Config{
		Vhost:     mqConf.Vhosts,
		Heartbeat: 10 * time.Second,
		Locale:    "zh_CN",
	})
	if err != nil {
		panic(err)
	}
	return conn
}

func (c *client) StoreClient(key string, value *amqp.Connection) {
	c.Store(key, value)
}

func (c *client) GetClient(key string) (*amqp.Connection, bool) {
	if value, exist := c.Load(key); exist {
		if cc, ok := value.(*amqp.Connection); ok && !cc.IsClosed() {
			return cc, true
		}
		return c.ReConnect(key)
	}
	return nil, false
}

func (c *client) ReConnect(key string) (*amqp.Connection, bool) {
	cc, _, _ := g.Do(key, func() (interface{}, error) {
		if conf, ok := configMap[key]; ok {
			cc := getRabbitMqConn(conf)
			c.Store(key, cc)
			return cc, nil
		}
		return nil, errors.New("not match map")
	})
	conn, ok := cc.(*amqp.Connection)
	if !ok {
		return nil, false
	}
	return conn, true
}
