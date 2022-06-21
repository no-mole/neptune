package model

type Config struct {
	Type string `yaml:"type"`

	Endpoint string `yaml:"endpoint"` //环境变量覆盖

	Settings map[string]string `yaml:"settings"`
	UserName string            `yaml:"username"`
	Password string            `yaml:"password"`
}
