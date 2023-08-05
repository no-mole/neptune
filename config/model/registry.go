package model

// Registry SC information
type Registry struct {
	Type string `yaml:"type"` //默认etcd

	Endpoint string `yaml:"endpoint"` //环境变量覆盖
	UserName string `yaml:"username"`
	Password string `yaml:"password"`

	APIVersion string //api版本
}
