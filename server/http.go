package server

import (
	"context"
	"errors"
	"github.com/no-mole/neptune/application"
	"github.com/no-mole/neptune/logger"
	"gopkg.in/yaml.v3"
	"net"
	"net/http"
)

func NewHttpServerPlugin(handlerFn func(ctx context.Context) http.Handler) application.Plugin {
	plg := &HttpServerPlugin{
		Plugin: application.NewPluginConfig("http-server", &application.PluginConfigOptions{
			ConfigName: "app.yaml",
			ConfigType: "yaml",
			EnvPrefix:  "",
		}),
		handlerFn: handlerFn,
		conf:      &HttpServerPluginConf{},
	}
	plg.Flags().StringVar(&plg.conf.Endpoint, "http-endpoint", "0.0.0.0:80", "http server endpoint,default is [0.0.0.0:80]")
	return plg
}

type HttpServerPlugin struct {
	application.Plugin `yaml:"-" json:"-"`

	handlerFn func(ctx context.Context) http.Handler
	handler   http.Handler
	listener  net.Listener

	conf *HttpServerPluginConf
}

type HttpServerPluginConf struct {
	Endpoint string `json:"http-endpoint" yaml:"http-endpoint"`
}

var ErrorEmptyHttpEndpoint = errors.New("http server plugin used but not initialization")

func (h *HttpServerPlugin) Config(_ context.Context, conf []byte) error {
	return yaml.Unmarshal(conf, h.conf)
}

func (h *HttpServerPlugin) Init(ctx context.Context) error {
	logger.Info(
		ctx,
		"http server init",
		logger.WithField("httpServerEndpoint", h.conf.Endpoint),
	)
	if h.conf.Endpoint == "" {
		return ErrorEmptyHttpEndpoint
	}
	host, port, err := net.SplitHostPort(h.conf.Endpoint)
	if err != nil {
		return err
	}
	h.listener, err = net.Listen("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return err
	}
	return nil
}
func (h *HttpServerPlugin) Run(ctx context.Context) error {
	h.handler = h.handlerFn(ctx)
	logger.Info(
		ctx,
		"http server started",
		logger.WithField("httpServerEndpoint", h.conf.Endpoint),
	)
	return http.Serve(h.listener, h.handler)
}
