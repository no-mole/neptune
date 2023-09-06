package config

import (
	"context"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	EtcdDefaultDailTimeout          = time.Second
	EtcdDefaultDialKeepAliveTime    = 3 * time.Second
	EtcdDefaultDialKeepAliveTimeout = 3 * time.Second
)

func Trans2EtcdConfig(ctx context.Context, config *Config) clientv3.Config {
	etcdConf := clientv3.Config{
		Context:   ctx,
		Endpoints: strings.Split(config.Endpoints, ","),
		Username:  config.Username,
		Password:  config.Password,
		LogConfig: nil,
	}
	if val, ok := config.Settings["dial_timeout"]; ok {
		if dialTimeout, err := strconv.Atoi(val); err == nil {
			etcdConf.DialTimeout = time.Duration(dialTimeout) * time.Second
		}
	} else {
		etcdConf.DialTimeout = EtcdDefaultDailTimeout
	}

	if val, ok := config.Settings["dial_keepalive_time"]; ok {
		if dialKeepaliveTime, err := strconv.Atoi(val); err == nil {
			etcdConf.DialKeepAliveTime = time.Duration(dialKeepaliveTime) * time.Second
		}
	} else {
		etcdConf.DialKeepAliveTime = EtcdDefaultDialKeepAliveTime
	}

	if val, ok := config.Settings["dial_keepalive_timeout"]; ok {
		if dialKeepaliveTimeout, err := strconv.Atoi(val); err == nil {
			etcdConf.DialKeepAliveTimeout = time.Duration(dialKeepaliveTimeout) * time.Second
		}
	} else {
		etcdConf.DialKeepAliveTimeout = EtcdDefaultDialKeepAliveTimeout
	}

	return etcdConf
}

func Trans2NacosConfig(_ context.Context, config *Config) (clientConfig *constant.ClientConfig, serverConfigs []constant.ServerConfig, err error) {
	for _, ep := range strings.Split(config.Endpoints, ",") {
		scheme, host, port, path, err := getHostAndPort(ep)
		if err != nil {
			return nil, nil, err
		}
		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr:      host,
			ContextPath: path,
			Port:        port,
			Scheme:      scheme,
		})
	}
	//create ClientConfig
	clientConfig = constant.NewClientConfig(
		constant.WithUsername(config.Username),
		constant.WithPassword(config.Password),
		constant.WithNamespaceId(config.Namespace),
		constant.WithTimeoutMs(2000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogLevel("warn"),
	)
	if logDir, ok := config.Settings["log_dir"]; ok {
		clientConfig.LogDir = logDir
	}
	if cacheDir, ok := config.Settings["cache_dir"]; ok {
		clientConfig.CacheDir = cacheDir
	}
	if logLevel, ok := config.Settings["log_level"]; ok {
		clientConfig.LogLevel = logLevel
	}
	return
}
func getHostAndPort(rawurl string) (scheme, host string, port uint64, path string, err error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", "", 0, "", err
	}
	host = u.Hostname()
	portStr := u.Port()
	if portStr != "" {
		portInt, err := strconv.ParseUint(portStr, 10, 64)
		if err != nil {
			return "", "", 0, "", err
		}
		port = portInt
	} else if u.Scheme == "http" {
		port = 80
	} else if u.Scheme == "https" {
		port = 443
	}
	return u.Scheme, host, port, u.Path, nil
}
