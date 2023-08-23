package elastic_search

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/olivere/elastic/v7/trace/opentelemetry"
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
	Username string `json:"username" yaml:"username" validate:"required"`
	Password string `json:"password" yaml:"password" validate:"required"`
}

func InitElasticSearch(esName string, confStr string, opts ...elastic.ClientOptionFunc) error {
	esConf := &Config{}
	err := json.Unmarshal([]byte(confStr), esConf)
	if err != nil {
		panic(err)
	}

	httpClient := &http.Client{
		Transport: opentelemetry.NewTransport(),
	}

	options := []elastic.ClientOptionFunc{
		elastic.SetURL(esConf.Host),
		elastic.SetBasicAuth(esConf.Username, esConf.Password),
		elastic.SetSniff(false),
		elastic.SetHealthcheckTimeoutStartup(5 * time.Second),
		elastic.SetHealthcheckTimeout(1 * time.Second),
		elastic.SetHealthcheckInterval(60 * time.Second),
		elastic.SetHttpClient(httpClient),
	}

	if len(opts) > 0 {
		options = append(options, opts...)
	}
	client, err := elastic.NewClient(options...)
	if err != nil {
		panic(err)
	}

	Client.StoreClient(esName, client)
	return nil
}

func (c *client) StoreClient(key string, value *elastic.Client) {
	c.Store(key, value)
}

func (c *client) GetClient(key string) (*elastic.Client, bool) {
	if value, exist := c.Load(key); exist {
		if cc, ok := value.(*elastic.Client); ok {
			return cc, true
		}
	}
	return nil, false
}
