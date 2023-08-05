package center

import (
	"context"
	"fmt"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const defaultEtcdEndpoint = "127.0.0.1:2379"

type Etcd struct {
	opt     *config
	client  *clientv3.Client
	closeCh chan struct{}
}

func init() {
	RegistryImplementation("etcd", &Etcd{})
	RegistryImplementation("nacos", &NaCos{})

}

func (s *Etcd) Init(opts ...Option) error {
	s.opt = ApplyOptions(opts...)

	if s.opt.Endpoint == "" {
		s.opt.Endpoint = defaultEtcdEndpoint
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(s.opt.Endpoint, ","),
		DialTimeout: 5 * time.Second,
		Username:    s.opt.Auth.Username,
		Password:    s.opt.Auth.Password,
	})
	if err != nil {
		return err
	}
	s.client = cli
	return nil
}

func (s *Etcd) Close() error {
	close(s.closeCh)
	return s.client.Close()
}
func (s *Etcd) Set(ctx context.Context, key, value string) error {
	_, err := s.client.Put(ctx, s.genKey(key), value)
	return err
}
func (s *Etcd) SetEX(ctx context.Context, key, value string, ttl int64) error {
	leaseResp, err := s.client.Grant(ctx, ttl)
	if err != nil {
		return err
	}
	_, err = s.client.Put(ctx, s.genKey(key), value, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}
	return err
}
func (s *Etcd) SetExKeepAlive(ctx context.Context, key, value string, ttl int64) error {
	leaseResp, err := s.client.Grant(ctx, ttl)
	if err != nil {
		return err
	}
	_, err = s.client.Put(ctx, s.genKey(key), value, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}
	_, err = s.client.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		return err
	}
	return nil
}
func (s *Etcd) Get(ctx context.Context, key string) (*Item, error) {
	resp, err := s.client.Get(ctx, s.genKey(key))
	if err != nil {
		return nil, err
	}
	value := ""
	if len(resp.Kvs) == 0 {
		value = ""
	} else {
		value = string(resp.Kvs[0].Value)
	}
	return &Item{
		Namespace: s.opt.Namespace,
		Key:       key,
		value:     value,
		IsDefault: false,
	}, nil
}
func (s *Etcd) GetDefault(ctx context.Context, key string, defaultValue string) (*Item, error) {
	item, err := s.Get(ctx, s.genKey(key))
	if err != nil {
		return nil, err
	}
	if item.value == "" {
		item.value = defaultValue
		item.IsDefault = true
	}
	return item, nil
}

func (s *Etcd) GetWithPrefixKey(ctx context.Context, prefixKey string) (*Item, error) {
	resp, err := s.client.Get(ctx, s.genKey(prefixKey), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	kvs := make([]*KVs, 0)
	for _, v := range resp.Kvs {
		kv := &KVs{
			Key:   string(v.Key),
			Value: string(v.Value),
		}
		kvs = append(kvs, kv)
	}
	return &Item{
		Namespace: s.opt.Namespace,
		Key:       prefixKey,
		Kvs:       kvs,
		IsDefault: false,
	}, nil
}

func (s *Etcd) Exist(ctx context.Context, key string) (bool, error) {
	return true, nil //todo
}

func (s *Etcd) Watch(ctx context.Context, item *Item, callback func(item *Item)) {
	if callback == nil {
		return
	}
	watchCh := s.client.Watch(ctx, s.genKey(item.Key))
	go func() {
		for {
			select {
			case wResp := <-watchCh:
				if len(wResp.Events) > 0 {
					event := wResp.Events[len(wResp.Events)-1]
					item.SetValue(string(event.Kv.Value))
					callback(item)
				}
			case _, ok := <-s.closeCh:
				if !ok {
					return
				}
			}
		}
	}()
}

func (s *Etcd) WatchWithPrefix(ctx context.Context, item *Item, callback func(item *Item)) {
	if callback == nil {
		return
	}
	watchCh := s.client.Watch(ctx, s.genKey(item.Key), clientv3.WithPrefix())
	go func() {
		for {
			select {
			case wResp := <-watchCh:
				if len(wResp.Events) > 0 {
					event := wResp.Events[len(wResp.Events)-1]
					item.Act = int64(event.Type)
					item.Key = string(event.Kv.Key)
					item.SetValue(string(event.Kv.Value))
					callback(item)
				}
			case _, ok := <-s.closeCh:
				if !ok {
					return
				}
			}
		}
	}()
}

var _ Client = &Etcd{} //ensure Etcd Implementation Client

func (s *Etcd) genKey(key string) string {
	return fmt.Sprintf("/%s/%s", s.opt.Namespace, key)
}
