package database

import (
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

func init() {
	RegisterPlugin("otlp", Otlp)
}

func Otlp(conf *Config) gorm.Plugin {
	return tracing.NewPlugin(
		tracing.WithoutMetrics(),
		tracing.WithAttributes(
			attribute.String("database_driver", conf.Driver),
			attribute.String("database_host", conf.Host),
			attribute.Int("database_port", conf.Port),
			attribute.String("database_db_name", conf.Database),
			attribute.String("database_username", conf.Username),
		),
	)
}
