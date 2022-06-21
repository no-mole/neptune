package mongo

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/no-mole/neptune/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	conns map[string]*BaseMongo
)

func init() {
	structCodec, _ := bsoncodec.NewStructCodec(bsoncodec.JSONFallbackStructTagParser)
	bson.DefaultRegistry = bson.NewRegistryBuilder().RegisterDefaultEncoder(reflect.Struct, structCodec).RegisterDefaultDecoder(reflect.Struct, structCodec).Build()

	conns = make(map[string]*BaseMongo)
}

type MongoConfig struct {
	ReplicaSetName string `json:"replica_set_name" yaml:"replica_set_name"`
	AppUrl         string `json:"app_url" yaml:"app_url"  validate:"required"`
	Database       string `json:"database" yaml:"database"  validate:"required"`
	Username       string `json:"username" yaml:"username"  validate:"required"`
	Password       string `json:"password" yaml:"password"  validate:"required"`
	MaxPoolSize    int64  `json:"max_pool_size" yaml:"max_pool_size"  validate:"required"`
	MaxIdleTimeMS  int64  `json:"max_idle_time_ms" yaml:"max_idle_time_ms"  validate:"required"`
}

type BaseMongo struct {
	*mongo.Client
	DataBase string
}

func (b *BaseMongo) SetClient(engine string) bool {
	if client, ok := conns[engine]; ok {
		b.DataBase = client.DataBase
		b.Client = client.Client
		return true
	}
	return false
}

func InitMonClient(ctx context.Context, mongoName, strCon string) error {
	mongoConf := new(MongoConfig)
	err := json.Unmarshal([]byte(strCon), &mongoConf)
	if err != nil {
		panic(err)
	}

	baseMongo := new(BaseMongo)
	baseMongo.Client = InitMongoConn(ctx, mongoConf)
	baseMongo.DataBase = mongoConf.Database
	conns[mongoName] = baseMongo
	return nil
}

func InitMongoConn(ctx context.Context, cfg *MongoConfig) *mongo.Client {
	credential := options.Credential{
		Username:   cfg.Username,
		Password:   cfg.Password,
		AuthSource: cfg.Database,
	}

	client, err := mongo.Connect(ctx, options.Client().
		ApplyURI(fmt.Sprintf("mongodb://%s", cfg.AppUrl)).
		SetAuth(credential).
		SetMaxPoolSize(uint64(cfg.MaxPoolSize)).
		SetReplicaSet(cfg.ReplicaSetName).
		SetMaxConnIdleTime(time.Duration(cfg.MaxIdleTimeMS)).
		SetRegistry(bson.DefaultRegistry))
	if err != nil {
		panic(err)
	}
	return client
}
