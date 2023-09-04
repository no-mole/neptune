package server

import (
	"context"
	"errors"
	"github.com/no-mole/neptune/application"
	"github.com/no-mole/neptune/grpc_service"
	"github.com/no-mole/neptune/json"
	"github.com/no-mole/neptune/utils"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"net"
)

type GrpcService struct {
	Metadata grpc_service.Metadata
	Impl     any
}

func NewGrpcServerPlugin(svr *grpc.Server, services ...GrpcService) application.Plugin {
	plg := &GrpcServicePlugin{
		Plugin: application.NewPluginConfig("grpc-server", &application.PluginConfigOptions{
			ConfigName: "app",
			ConfigType: "yaml",
			EnvPrefix:  "grpc",
		}),
		services: services,
		server:   svr,
		err:      make(chan error, 1),
	}
	for _, service := range services {
		svr.RegisterService(service.Metadata.ServiceDesc(), service.Impl)
	}
	plg.Command().PersistentFlags().StringVar(&plg.GrpcListen, "grpc-endpoint", "0.0.0.0:8081", "grpc监听地址,默认为 [0.0.0.0:8081]")
	plg.Command().PersistentFlags().StringVar(&plg.ServiceEndpoint, "service-endpoint", "", "服务注册使用的地址，从环境变量中取或者取第一个非回环ip [ip:port]")
	return plg
}

type GrpcServicePlugin struct {
	application.Plugin `yaml:"-"`

	server   *grpc.Server `yaml:"-"`
	listener net.Listener `yaml:"-"`

	services []GrpcService `yaml:"-"`

	GrpcListen      string `yaml:"grpc_listen" json:"grpc_listen"`
	ServiceEndpoint string `yaml:"service_endpoint" json:"service_endpoint"`
	ep              string `yaml:"-"`

	err chan error `yaml:"-"`
}

var ErrorEmptyEndpoint = errors.New("grpc server plugin used but not initialization")

func (g *GrpcServicePlugin) Init(_ context.Context, conf []byte) error {
	err := json.Unmarshal(conf, g)
	if err != nil {
		return err
	}
	g.Command().RunE = func(cmd *cobra.Command, args []string) error {
		ep, err := g.DiscoverTheEntrance()
		if err != nil {
			return err
		}
		g.ep = ep

		if g.GrpcListen == "" {
			return ErrorEmptyEndpoint
		}
		host, port, err := net.SplitHostPort(g.GrpcListen)
		if err != nil {
			return err
		}

		listener, err := net.Listen("tcp", net.JoinHostPort(host, port))
		if err != nil {
			return err
		}
		g.listener = listener
		go func() {
			g.err <- g.server.Serve(g.listener)
		}()
		return nil
	}
	err = g.Command().Execute()
	if err != nil {
		return err
	}

	g.Command().RunE = func(cmd *cobra.Command, args []string) error {
		for _, service := range g.services {
			err := grpc_service.Register(context.Background(), g.ep, service.Metadata)
			if err != nil {
				return err
			}
		}
		defer func() {
			close(g.err)
		}()
		return <-g.err
	}
	return nil
}

func (g *GrpcServicePlugin) DiscoverTheEntrance() (string, error) {
	if g.ServiceEndpoint == "" {
		g.ServiceEndpoint = g.GrpcListen
	}
	host, port, err := net.SplitHostPort(g.ServiceEndpoint)
	if err != nil {
		return "", err
	}
	if host == "" || net.ParseIP(host).Equal(net.IPv4zero) {
		curHost, err := utils.GetSystemIP()
		if err != nil {
			return "", err
		}
		host = curHost
	}
	return net.JoinHostPort(host, port), nil
}
