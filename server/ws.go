package server

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/no-mole/neptune/application"
	"github.com/no-mole/neptune/logger"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"net"
	"net/http"
)

func NewWebSocketServerPlugin(handlerFn func(ctx context.Context, w http.ResponseWriter, r *http.Request)) application.Plugin {
	plg := &WebSocketServerPlugin{
		Plugin: application.NewPluginConfig("ws-server", &application.PluginConfigOptions{
			ConfigName: "app.yaml",
			ConfigType: "yaml",
			EnvPrefix:  "",
		}),
		handlerFn: handlerFn,
		conf:      &WsServerPluginConf{},
	}
	plg.Flags().StringVar(&plg.conf.Endpoint, "ws-endpoint", "0.0.0.0:80", "ws server endpoint,default is [0.0.0.0:80]")
	return plg
}

type WebSocketServerPlugin struct {
	application.Plugin `yaml:"-" json:"-"`

	handlerFn func(ctx context.Context, w http.ResponseWriter, r *http.Request) // WebSocket 处理函数
	server    *http.Server
	listener  net.Listener
	upgrader  websocket.Upgrader

	conf *WsServerPluginConf
}

type WsServerPluginConf struct {
	Endpoint string `json:"ws-endpoint" yaml:"ws-endpoint"`
}

var ErrorEmptyWsEndpoint = errors.New("ws server plugin used but not initialization")

func (w *WebSocketServerPlugin) Config(_ context.Context, conf []byte) error {
	return yaml.Unmarshal(conf, w.conf)
}

func (w *WebSocketServerPlugin) Init(ctx context.Context) error {
	logger.Info(
		ctx,
		"webSocket server init",
		logger.WithField("wsServerEndpoint", w.conf.Endpoint),
	)
	if w.conf.Endpoint == "" {
		return ErrorEmptyWsEndpoint
	}
	host, port, err := net.SplitHostPort(w.conf.Endpoint)
	if err != nil {
		return err
	}
	w.listener, err = net.Listen("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return err
	}
	// 设置 WebSocket 升级器
	w.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许任何来源
		},
	}

	// 创建 http.Server 实例
	w.server = &http.Server{
		Addr: w.conf.Endpoint,
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			w.handlerFn(ctx, rw, r) // 调用用户提供的 WebSocket 处理函数
		}),
	}

	return nil
}
func (w *WebSocketServerPlugin) Run(ctx context.Context) error {

	logger.Info(
		ctx,
		"webSocket server started",
		logger.WithField("wsServerEndpoint", w.conf.Endpoint),
	)
	// 运行 WebSocket 服务器
	return w.server.Serve(w.listener)
}
