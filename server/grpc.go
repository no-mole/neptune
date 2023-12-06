package server

import (
	"context"
	"errors"
	"net"

	"github.com/no-mole/neptune/application"
	"github.com/no-mole/neptune/grpc_service"
	"github.com/no-mole/neptune/logger"
	"github.com/no-mole/neptune/utils"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

type GrpcService struct {
	Metadata grpc_service.Metadata
	Impl     any
}

func NewGrpcServerPlugin(grpcServerFn func(ctx context.Context) *grpc.Server, services ...GrpcService) application.Plugin {
	plg := &GrpcServerPlugin{
		Plugin: application.NewPluginConfig("grpc-server", &application.PluginConfigOptions{
			ConfigName: "app.yaml",
			ConfigType: "yaml",
			EnvPrefix:  "",
		}),
		fn:       grpcServerFn,
		services: services,
		err:      make(chan error, 1),
		conf:     &GrpcServerPluginConf{},
	}
	plg.Flags().StringVar(&plg.conf.GrpcListen, "grpc-endpoint", "0.0.0.0:8080", "grpc监听地址,默认为 [0.0.0.0:8080]")
	plg.Flags().StringVar(&plg.conf.ServiceEndpoint, "service-endpoint", "", "服务注册使用的地址，从环境变量中取或者取第一个非回环ip [ip:port]")
	return plg
}

type GrpcServerPlugin struct {
	application.Plugin `yaml:"-"`

	fn       func(ctx context.Context) *grpc.Server
	server   *grpc.Server `yaml:"-"`
	listener net.Listener `yaml:"-"`

	services []GrpcService `yaml:"-"`

	ep string `yaml:"-"`

	err chan error `yaml:"-"`

	conf *GrpcServerPluginConf
}

type GrpcServerPluginConf struct {
	GrpcListen      string `yaml:"grpc-endpoint" json:"grpc-endpoint"`
	ServiceEndpoint string `yaml:"service_endpoint" json:"service_endpoint"`
}

var ErrorEmptyEndpoint = errors.New("grpc server plugin used but not initialization")

func (g *GrpcServerPlugin) Config(_ context.Context, conf []byte) error {
	return yaml.Unmarshal(conf, g.conf)
}

func (g *GrpcServerPlugin) Init(ctx context.Context) error {
	ep, err := g.DiscoverTheEntrance()
	if err != nil {
		logger.Error(
			ctx,
			"grpc server init discover the entrance",
			err,
			logger.WithField("grpcServerListen", g.conf.GrpcListen),
			logger.WithField("grpcServiceEndpoint", g.conf.ServiceEndpoint),
		)
		return err
	}
	logger.Info(
		ctx,
		"grpc server init",
		logger.WithField("grpcServerListen", g.conf.GrpcListen),
		logger.WithField("grpcServiceEndpoint", g.conf.ServiceEndpoint),
		logger.WithField("grpcServiceEntrance", g.ep),
	)
	g.ep = ep
	if g.conf.GrpcListen == "" {
		return ErrorEmptyEndpoint
	}
	host, port, err := net.SplitHostPort(g.conf.GrpcListen)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return err
	}
	g.listener = listener
	return nil
}
func (g *GrpcServerPlugin) Run(ctx context.Context) error {
	g.server = g.fn(ctx)
	for _, service := range g.services {
		g.server.RegisterService(service.Metadata.ServiceDesc(), service.Impl)
	}
	go func() {
		g.err <- g.server.Serve(g.listener)
	}()
	for _, service := range g.services {
		err := grpc_service.Register(context.Background(), g.ep, service.Metadata)
		if err != nil {
			return err
		}
	}
	defer func() {
		close(g.err)
	}()
	logger.Info(
		ctx,
		"grpc server started",
		logger.WithField("grpcServerListen", g.conf.GrpcListen),
		logger.WithField("grpcServiceEndpoint", g.conf.ServiceEndpoint),
		logger.WithField("grpcServiceEntrance", g.ep),
	)
	select {
	case <-ctx.Done():
		return nil
	case err := <-g.err:
		return err
	}
}

func (g *GrpcServerPlugin) DiscoverTheEntrance() (string, error) {
	if g.conf.ServiceEndpoint == "" {
		g.conf.ServiceEndpoint = g.conf.GrpcListen
	}
	host, port, err := net.SplitHostPort(g.conf.ServiceEndpoint)
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
