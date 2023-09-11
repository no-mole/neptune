package application

import (
	"context"
	"fmt"
	"github.com/no-mole/neptune/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type HookFunc func(ctx context.Context) error

type AppMode string

var (
	AppModeProd = "prod"
	AppModeGrey = "grey"
	AppModeTest = "test"
	AppModeDev  = "dev"

	DefaultEnvPrefix = "neptune"
)

type App struct {
	command *cobra.Command

	ctx    context.Context
	cancel context.CancelFunc

	plugins []Plugin
	hooks   []HookFunc

	Mode      string //app run mode,default is [prod]
	LogLevel  string //slog level,default is info,[trace|debug|notice|info|warn|error|fatal]
	EnvPrefix string //app global env prefix,default is neptune
}

func New(ctx context.Context) *App {
	newCtx, cancel := context.WithCancel(ctx)

	app := &App{
		ctx:       newCtx,
		cancel:    cancel,
		EnvPrefix: DefaultEnvPrefix,
	}

	app.command = &cobra.Command{
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			//init app config for flags and env
			err := app.initConfig()
			if err != nil {
				return err
			}
			//init logger
			conf := zap.NewProductionConfig()
			conf.EncoderConfig.CallerKey = ""
			conf.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
			zapLogger, _ := conf.Build()
			newLogger := logger.NewLogger(app.ctx, zapLogger)
			newLogger.SetLevel(logger.GetLevelByName(app.LogLevel))
			logger.SetLogger(newLogger)
			logger.Info(app.ctx, "application pre run", logger.WithField("mode", app.Mode), logger.WithField("envPrefix", app.EnvPrefix))
			//plugin init
			for _, plg := range app.plugins {
				err = app.pluginConfigInit(plg)
				if err != nil {
					return err
				}
				err = app.pluginEnvBind(plg)
				if err != nil {
					return err
				}
				err = plg.Init(app.ctx)
				if err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			//run hooks
			for _, hook := range app.hooks {
				err = hook(ctx)
				if err != nil {
					logger.Error(app.ctx, "run hook err", err)
					return err
				}
			}
			plgNames := make([]string, 0, len(app.plugins))
			for _, plg := range app.plugins {
				plgNames = append(plgNames, plg.Name())
			}
			logger.Info(app.ctx, "application will started", logger.WithField("mode", app.Mode), logger.WithField("envPrefix", app.EnvPrefix), logger.WithField("plugins", plgNames))
			eg, _ := errgroup.WithContext(app.ctx)
			for _, plg := range app.plugins {
				eg.Go(
					func(p Plugin) func() error {
						return func() error {
							return p.Run(app.ctx)
						}
					}(plg))
			}
			defer func() {
				if err != nil {
					logger.Error(app.ctx, "app shutdown", err)
				} else {
					logger.Info(app.ctx, "app shutdown")
				}
			}()
			return eg.Wait()
		},
	}
	app.command.PersistentFlags().StringVar(&app.Mode, "mode", AppModeDev, "app run mode for [prod|grey|test|dev],default is dev")
	app.command.PersistentFlags().StringVar(&app.LogLevel, "log-level", logger.LevelInfo.String(), "slog level,default is info,[debug|info|warn|error]")
	return app
}

func (app *App) Run() error {
	app.listenSigns()
	for _, plg := range app.plugins {
		app.command.PersistentFlags().AddFlagSet(plg.Flags())
	}
	return app.command.Execute()
}

func (app *App) Cancel(msg string, args ...any) {
	app.cancel()
	logger.Info(app.ctx, fmt.Sprintf(msg, args...))
}

// Use Plugins running after server startup
func (app *App) Use(plugins ...Plugin) {
	app.plugins = append(app.plugins, plugins...)
}

func (app *App) Hook(hooks ...HookFunc) {
	app.hooks = append(app.hooks, hooks...)
}

func (app *App) initConfig() error {
	v := viper.New()
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.SetConfigName("app.yaml")
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err == nil {
		logger.Info(app.ctx, "using config file", logger.WithField("configFileUsed", v.ConfigFileUsed()))
	}
	v.SetEnvPrefix(app.EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()
	BindFlags(app.ctx, "app", app.command.Flags(), v)
	return nil
}

func (app *App) listenSigns() {
	go func() {
		signs := make(chan os.Signal, 1)
		signal.Notify(signs, syscall.SIGKILL, syscall.SIGTERM)
		select {
		case <-app.ctx.Done():
		case sign := <-signs:
			app.Cancel("receive sig:[%s],app canceled", sign.String())
		}
		signal.Stop(signs)
		close(signs)
	}()
}

func (app *App) pluginConfigInit(plg Plugin) error {
	v := viper.New()
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath(fmt.Sprintf("./config/%s", app.Mode))

	v.SetConfigFile(plg.ConfigOptions().ConfigFile)
	v.SetConfigName(plg.ConfigOptions().ConfigName)
	v.SetConfigType(plg.ConfigOptions().ConfigType)

	opts := []zap.Field{logger.WithField("pluginName", plg.Name()), logger.WithField("pluginConfigOpts", plg.ConfigOptions())}

	if err := v.ReadInConfig(); err != nil {
		logger.Info(app.ctx, "init plugin with no config file", opts...)
		//using no config file
		return nil
	}
	opts = append(opts, logger.WithField("configFileUsed", v.ConfigFileUsed()))
	logger.Info(app.ctx, "init plugin", opts...,
	)
	//优先使用配置文件初始化
	body, err := os.ReadFile(v.ConfigFileUsed())
	if err != nil {
		logger.Error(app.ctx, "init plugin error", err, opts...)
		return err
	}
	err = plg.Config(app.ctx, body)
	if err != nil {
		logger.Error(app.ctx, "init plugin error", err, opts...)
	}
	return err
}

func (app *App) pluginEnvBind(plg Plugin) error {
	v := viper.New()
	//其次配置环境变量前缀 ${app_env_prefix}_${plugin_env_prefix}
	envPrefix := app.EnvPrefix
	if plg.ConfigOptions().EnvPrefix != "" {
		envPrefix += "_" + plg.ConfigOptions().EnvPrefix
	}
	logger.Debug(app.ctx, "init plugin", logger.WithField("pluginName", plg.Name()), logger.WithField("pluginEnvPrefix", envPrefix))
	v.SetEnvPrefix(envPrefix)
	// Environment variables can't have dashes in them, so bind them to their equivalent
	// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()
	//环境变量注入flag
	BindFlags(app.ctx, plg.Name(), plg.Flags(), v)
	return nil
}

// BindFlags Bind each cobra flag to its associated viper configuration (config file and environment variable)
func BindFlags(ctx context.Context, groupName string, set *pflag.FlagSet, v *viper.Viper) {
	call := func(f *pflag.Flag) {
		// Determine the naming convention of the flags when represented in the config file
		configName := f.Name
		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			err := set.Set(f.Name, fmt.Sprintf("%v", val))
			if err != nil {
				logger.Error(ctx, "bind flag error", err, logger.WithField("groupName", groupName), logger.WithField("flagName", f.Name), logger.WithField("setVal", val))
			}
		}
	}
	set.VisitAll(call)
}
