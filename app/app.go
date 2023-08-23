package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/no-mole/neptune/config"
	"github.com/no-mole/neptune/registry"
	"github.com/no-mole/neptune/utils"
	//"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	_ "go.uber.org/automaxprocs"
	"google.golang.org/grpc"
)

type HookFunc func(ctx context.Context) error

var app *APP

func init() {
	app = NewApp(context.Background())
	go func() {
		signs := make(chan os.Signal, 1)
		signal.Notify(signs, syscall.SIGKILL, syscall.SIGTERM)
		sign := <-signs
		fmt.Printf("Listen System Signal:%s,App Stopping!", sign)
		Stop()
	}()
}

func NewApp(ctx context.Context) *APP {
	ctx, cancel := context.WithCancel(ctx)
	app = &APP{
		ctx:        ctx,
		cancel:     cancel,
		HttpEngine: gin.New(),
		GrpcEngine: grpc.NewServer(),
	}
	return app
}

type APP struct {
	ctx    context.Context
	cancel context.CancelFunc

	hooks []*hook

	errorCh chan error

	httpLister net.Listener
	HttpEngine *gin.Engine

	grpcLister net.Listener
	GrpcEngine *grpc.Server
}

type hook struct {
	delay bool
	fn    HookFunc
}

func NewGrpcServer(opts ...grpc.ServerOption) *grpc.Server {
	app.GrpcEngine = grpc.NewServer(opts...)
	return app.GrpcEngine
}

func AddHook(hooks ...HookFunc) {
	for _, fn := range hooks {
		addHook(fn, false)
	}
}

func AddDelayHook(hooks ...HookFunc) {
	for _, fn := range hooks {
		addHook(fn, true)
	}
}

func addHook(fn HookFunc, delay bool) {
	app.hooks = append(app.hooks, &hook{
		delay: delay,
		fn:    fn,
	})
}

func Start() error {
	AddHook(startHttpServe)
	AddHook(startGrpcServe)
	flag := true
	for i := 0; i < 2; i++ {
		for _, hook := range app.hooks {
			if hook.delay == flag {
				continue
			}
			if err := hook.fn(app.ctx); err != nil {
				stack := utils.GetStack(2, 16)
				fnName := runtime.FuncForPC(reflect.ValueOf(hook.fn).Pointer()).Name()
				return fmt.Errorf("app start err!\nhookName:%s\nerror:%s\nstack:\n%s", fnName, err.Error(), stack)
			}
		}
		flag = false
	}
	return nil
}

func startGrpcServe(ctx context.Context) (err error) {
	if config.GlobalConfig.GrpcPort == 0 {
		return nil
	}
	if registry.GetRegister() == nil {
		conf := config.GetRegistryConf()
		errCh := make(chan error, 1)
		switch conf.Type {
		case "etcd":
			reg, err := registry.NewEtcdRegister(ctx, &registry.EtcdRegisterConfig{
				Endpoints: conf.Endpoint,
				Username:  conf.UserName,
				Password:  conf.Password,
			}, errCh)
			if err != nil {
				return err
			}
			registry.SetRegister(reg)
		case "nacos":
			reg, err := registry.NewNaCosRegister(ctx, &registry.NaCosConfig{
				Username:  conf.UserName,
				Password:  conf.Password,
				Endpoint:  conf.Endpoint,
				Namespace: config.GlobalConfig.Namespace,
			}, errCh)
			if err != nil {
				return err
			}
			registry.SetRegister(reg)
		}

		go func() {
			//listen register errors
			err := <-errCh
			close(errCh)
			app.errorCh <- err
		}()
	}

	listen := fmt.Sprintf("0.0.0.0:%d", config.GlobalConfig.GrpcPort)
	listener, err := net.Listen("tcp", listen)
	if err != nil {
		return nil
	}
	fmt.Printf("GRPC SERVER LISTEN ON >>>> %s\n", listen)
	app.grpcLister = listener

	go func() {
		err := app.GrpcEngine.Serve(app.grpcLister)
		if err != nil {
			app.errorCh <- err
		}
	}()
	return nil
}

func startHttpServe(ctx context.Context) error {
	if config.GlobalConfig.HttpPort == 0 {
		return nil
	}

	listen := fmt.Sprintf("0.0.0.0:%d", config.GlobalConfig.HttpPort)
	listener, err := net.Listen("tcp", listen)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Printf("HTTP LISTEN ON >>>> %s\n", listen)

	app.httpLister = listener
	go func() {
		err = http.Serve(app.httpLister, app.HttpEngine)
		if err != nil {
			app.errorCh <- err
		}
	}()
	return nil
}

func Stop() {
	app.cancel()
	app.GrpcEngine.GracefulStop()
	time.Sleep(3 * time.Second)
	fmt.Println("App Stopped!")
}

func ErrorCh() chan error {
	return app.errorCh
}

func Error(err error) {
	go func() {
		app.errorCh <- err
	}()
}
