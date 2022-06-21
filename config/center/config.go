package center

type config struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Auth      Auth   `json:"auth" yaml:"auth"`
	Endpoint  string `json:"endpoint" yaml:"endpoint"`
}

type Auth struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

type optionFunc func(*config)

func (of optionFunc) Apply(cfg *config) { of(cfg) }

type Option interface {
	Apply(*config)
}

func WithNamespace(namespace string) Option {
	return optionFunc(func(f *config) {
		f.Namespace = namespace
	})
}

func WithAuth(auth Auth) Option {
	return optionFunc(func(f *config) {
		f.Auth = auth
	})
}

func WithEndpoint(endpoint string) Option {
	return optionFunc(func(f *config) {
		f.Endpoint = endpoint
	})
}

func ApplyOptions(opts ...Option) *config {
	conf := &config{}
	for _, opt := range opts {
		opt.Apply(conf)
	}
	return conf
}
