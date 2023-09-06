package application

import (
	"context"
	"github.com/spf13/pflag"
)

type Plugin interface {
	Name() string
	Flags() *pflag.FlagSet
	ConfigOptions() *PluginConfigOptions
	Config(ctx context.Context, config []byte) error
	Init(ctx context.Context) error
	Run(ctx context.Context) error
}

type PluginConfigOptions struct {
	ConfigFile string
	ConfigName string
	ConfigType string
	EnvPrefix  string
}

func NewPluginConfig(name string, opts *PluginConfigOptions) Plugin {
	return &basicPlugin{
		name:  name,
		flags: pflag.NewFlagSet(name, pflag.ContinueOnError),
		opts:  opts,
	}
}

type basicPlugin struct {
	name  string
	flags *pflag.FlagSet
	opts  *PluginConfigOptions
}

func (b *basicPlugin) Config(_ context.Context, _ []byte) error {
	return nil
}

func (b *basicPlugin) Init(ctx context.Context) error {
	return nil
}

func (b *basicPlugin) Run(_ context.Context) error {
	return nil
}

func (b *basicPlugin) Name() string {
	return b.name
}

func (b *basicPlugin) Flags() *pflag.FlagSet {
	return b.flags
}

func (b *basicPlugin) ConfigOptions() *PluginConfigOptions {
	return b.opts
}
