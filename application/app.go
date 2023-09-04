package application

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type AppMode string

var (
	AppModeProd = "prod"
	AppModeGrey = "grey"
	AppModeTest = "test"
	AppModeDev  = "dev"

	DefaultEnvPrefix = "neptune"
	DefaultVersion   = "v1"
	DefaultAppName   = "app"
)

type App struct {
	command *cobra.Command

	ctx    context.Context
	cancel context.CancelFunc

	plugins []Plugin
	hooks   []HookFunc

	Name    string //app name,default is [ app]
	Version string //app version,default is [v1]
	Mode    string //app run mode,default is [prod]
	Debug   bool   //debug mode,default is [false]
}

func New(ctx context.Context) *App {
	newCtx, cancel := context.WithCancel(ctx)

	app := &App{
		ctx:    newCtx,
		cancel: cancel,
	}

	app.command = &cobra.Command{
		Version: DefaultVersion,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			app.listenSigns()
			defer func() {
				//reset version
				app.command.Version = app.Version
			}()
			for _, plg := range app.plugins {
				app.command.AddCommand(plg.Command())
			}
			//init app config for flags and env
			return app.initConfig()
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			//plugin init
			for _, plg := range app.plugins {
				err := app.initPlugin(plg)
				if err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			//run hooks
			for _, hook := range app.hooks {
				err := hook(ctx)
				if err != nil {
					return err
				}
			}
			eg, _ := errgroup.WithContext(app.ctx)
			for _, plg := range app.plugins {
				eg.Go(
					func(p Plugin) func() error {
						return func() error {
							return p.Command().Execute()
						}
					}(plg))
			}
			return eg.Wait()
		},
	}
	app.command.PersistentFlags().StringVar(&app.Name, "name", DefaultAppName, "app name")
	app.command.PersistentFlags().StringVar(&app.Mode, "mode", AppModeProd, "app run mode for [prod|grey|test|dev]")
	app.command.PersistentFlags().BoolVar(&app.Debug, "debug", false, "app debug for [true|false]")
	return app
}

func (app *App) Run() error {
	return app.command.Execute()
}

func (app *App) Cancel(msg string, args ...any) {
	app.cancel()
	fmt.Printf(msg, args...)
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
	v.SetConfigFile("app.yaml")
	if err := v.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
	v.SetEnvPrefix(DefaultEnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()
	BindFlags(app.command.Flags(), v)
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

func (app *App) initPlugin(plg Plugin) error {
	v := viper.New()

	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath(fmt.Sprintf("./config/%s", app.Mode))

	v.SetConfigFile(plg.ConfigOptions().ConfigFile)
	v.SetConfigName(plg.ConfigOptions().ConfigName)
	v.SetConfigType(plg.ConfigOptions().ConfigType)

	if err := v.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
		//优先使用配置文件初始化
		body, err := os.ReadFile(viper.ConfigFileUsed())
		if err != nil {
			return err
		}
		err = plg.Init(app.ctx, body)
		if err != nil {
			return err
		}
	}
	//其次配置环境变量前缀 ${app_env_prefix}_${plugin_env_prefix}
	envPrefix := DefaultEnvPrefix
	if plg.ConfigOptions().EnvPrefix != "" {
		envPrefix += "_" + plg.ConfigOptions().EnvPrefix
	}
	v.SetEnvPrefix(envPrefix)
	// Environment variables can't have dashes in them, so bind them to their equivalent
	// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()
	//环境变量注入flag
	BindFlags(plg.Command().PersistentFlags(), v)
	BindFlags(plg.Command().Flags(), v)
	return nil
}

// BindFlags Bind each cobra flag to its associated viper configuration (config file and environment variable)
func BindFlags(set *pflag.FlagSet, v *viper.Viper) {
	call := func(f *pflag.Flag) {
		// Determine the naming convention of the flags when represented in the config file
		configName := f.Name
		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			err := set.Set(f.Name, fmt.Sprintf("%v", val))
			if err != nil {
				//todo logger error
			}
		}
	}
	set.VisitAll(call)
}
