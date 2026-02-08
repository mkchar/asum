package engine

type Options struct {
	Config Config
}

type Option func(*Options)

func WithConfig(cfg Config) Option {
	return func(o *Options) {
		o.Config = cfg
	}
}
