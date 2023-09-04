package server

import (
	"context"
	"errors"
	"github.com/no-mole/neptune/application"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"net"
	"net/http"
)

func NewHttpServerPlugin(handler http.Handler) application.Plugin {
	plg := &HttpServerPlugin{
		Plugin: application.NewPluginConfig("http-server", &application.PluginConfigOptions{
			ConfigName: "app",
			ConfigType: "yaml",
			EnvPrefix:  "",
		}),
		handler: handler,
	}
	plg.Command().PersistentFlags().StringVar(&plg.Endpoint, "http-endpoint", "0.0.0.0:8080", "http server endpoint,default is [0.0.0.0:8080]")
	return plg
}

type HttpServerPlugin struct {
	application.Plugin

	handler  http.Handler
	listener net.Listener

	Endpoint string `json:"http_endpoint" yaml:"http_endpoint"`
}

var ErrorEmptyHttpEndpoint = errors.New("http server plugin used but not initialization")

func (h *HttpServerPlugin) Init(_ context.Context, conf []byte) error {
	err := yaml.Unmarshal(conf, h)
	if err != nil {
		return err
	}

	h.Command().RunE = func(cmd *cobra.Command, args []string) error {
		if h.Endpoint == "" {
			return ErrorEmptyHttpEndpoint
		}
		host, port, err := net.SplitHostPort(h.Endpoint)
		if err != nil {
			return err
		}
		h.listener, err = net.Listen("tcp", net.JoinHostPort(host, port))
		if err != nil {
			return err
		}
		return nil
	}

	h.Command().RunE = func(cmd *cobra.Command, args []string) error {
		return http.Serve(h.listener, h.handler)
	}
	return nil
}
