package registry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/no-mole/neptune/config"

	"github.com/no-mole/neptune/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdRegisterConfig struct {
	Endpoints string `json:"endpoints" yaml:"endpoints"`
	Username  string `json:"username" yaml:"username"`
	Password  string `json:"password" yaml:"password"`
}

func NewEtcdRegister(_ context.Context, conf *EtcdRegisterConfig, errCh chan error) (_ Registration, err error) {
	r := &EtcdRegister{
		errCh: errCh,
	}
	r.client, err = clientv3.New(clientv3.Config{
		Endpoints:            strings.Split(conf.Endpoints, ","),
		DialTimeout:          1 * time.Second,
		DialKeepAliveTime:    5 * time.Second,
		DialKeepAliveTimeout: 1 * time.Second,
		PermitWithoutStream:  true,
		Username:             conf.Username,
		Password:             conf.Password,
	})
	r.grpcMetas = make([]GrpcMeta, 0)
	r.intervalTime = 5 * time.Second
	return r, err
}

type EtcdRegister struct {
	//client etcd client
	client *clientv3.Client

	//leaseID use one lease
	leaseID clientv3.LeaseID

	once sync.Once

	errCh chan error

	grpcMetas []GrpcMeta

	intervalTime time.Duration
}

//Register instance EtcdRegister etcd
func (r *EtcdRegister) Register(ctx context.Context, meta GrpcMeta) (err error) {
	r.once.Do(func() {
		err = r.initGrantKeepAlive(ctx)
		if err != nil {
			logger.Error(ctx, "app", err, logger.WithField("msg", fmt.Sprintf("etcd init grant error：%s", err.Error())))
			return
		}
	})
	if err != nil {
		return err
	}
	r.grpcMetas = append(r.grpcMetas, meta)

	key := meta.GenKey()
	Endpoint := fmt.Sprintf("%s:%d", config.GlobalConfig.IP, config.GlobalConfig.GrpcPort)
	//Register the service in etcd
	_, err = r.client.Put(ctx, key, Endpoint, clientv3.WithLease(r.leaseID))
	if err != nil {
		logger.Error(ctx, "app", err, logger.WithField("msg", fmt.Sprintf("etcd register put error：%s", err.Error())))
	}
	return err
}

func (r *EtcdRegister) keepAlive(ctx context.Context) (err error) {
	ch, err := r.client.KeepAlive(ctx, r.leaseID)
	if err != nil {
		logger.Error(ctx, "app", err, logger.WithField("msg", fmt.Sprintf("etcd register keepalive error：%s", err.Error())))
		return
	}
	go func() {
		for {
			_, ok := <-ch
			if !ok {
				err = errors.New("etcd register keepalive error:keepalive channel closed")
				logger.Error(ctx, "app", err)
				r.reKeepAlive(ctx)
				if r.errCh != nil {
					r.errCh <- err
				}
				return
			}
		}
	}()
	return
}

func (r *EtcdRegister) reKeepAlive(ctx context.Context) {

initGrant:
	err := r.initGrantKeepAlive(ctx)
	if err != nil {
		logger.Error(ctx, "app", err, logger.WithField("msg", fmt.Sprintf("etcd init grant error：%s", err.Error())))
		<-time.After(r.intervalTime)
		goto initGrant
	}
putMateKey:
	for _, meta := range r.grpcMetas {
		key := meta.GenKey()
		Endpoint := fmt.Sprintf("%s:%d", config.GlobalConfig.IP, config.GlobalConfig.GrpcPort)
		//Register the service in etcd
		_, err = r.client.Put(ctx, key, Endpoint, clientv3.WithLease(r.leaseID))
		if err != nil {
			logger.Error(ctx, "app", err, logger.WithField("msg", fmt.Sprintf("etcd register put error：%s", err.Error())))
			<-time.After(r.intervalTime)
			goto putMateKey
		}
	}
}

func (r *EtcdRegister) UnRegister(ctx context.Context, meta GrpcMeta) (err error) {
	_, err = r.client.Delete(ctx, meta.GenKey())
	if err != nil {
		return err
	}
	return
}

func (r *EtcdRegister) initGrantKeepAlive(ctx context.Context) error {
	var resp *clientv3.LeaseGrantResponse
	resp, err := r.client.Grant(ctx, TTL)
	if err != nil {
		return err
	}
	r.leaseID = resp.ID
	err = r.keepAlive(ctx)
	if err != nil {
		return err
	}
	return nil
}
