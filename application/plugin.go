package application

import (
	"context"
	"github.com/spf13/cobra"
)

type Plugin interface {
	Command() *cobra.Command
	ConfigOptions() *PluginConfigOptions
	Init(ctx context.Context, config []byte) error
}

type PluginConfigOptions struct {
	ConfigFile string
	ConfigName string
	ConfigType string
	EnvPrefix  string
}

func NewPluginConfig(name string, opts *PluginConfigOptions) Plugin {
	return &basicPlugin{
		command: &cobra.Command{
			Use: name,
			//GroupID: name,
		},
		opts: opts,
	}
}

type basicPlugin struct {
	command *cobra.Command
	opts    *PluginConfigOptions
}

func (b *basicPlugin) Init(_ context.Context, _ []byte) error {
	return nil
}

func (b *basicPlugin) Command() *cobra.Command {
	return b.command
}

func (b *basicPlugin) ConfigOptions() *PluginConfigOptions {
	return b.opts
}
