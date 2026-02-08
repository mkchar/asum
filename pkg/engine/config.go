package engine

import "time"

type Env string

const (
	EnvDev  Env = "dev"
	EnvProd Env = "prod"
)

type Config struct {
	AppName string
	Env     Env

	Host string
	Port int

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	BodyLimit   int
	Concurrency int

	CaseSensitive bool
	StrictRouting bool

	Immutable          bool
	UnescapePath       bool
	DisableKeepalive   bool
	DisableDefaultDate bool
	ServerHeader       string

	ShutdownTimeout time.Duration
}

func DefaultConfig() Config {
	return Config{
		AppName: "gateway",
		Env:     EnvDev,

		Host: "0.0.0.0",
		Port: 8080,

		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,

		BodyLimit:   8 * 1024 * 1024,
		Concurrency: 256 * 1024,

		CaseSensitive: false,
		StrictRouting: false,

		Immutable:          true,
		UnescapePath:       false,
		DisableKeepalive:   false,
		DisableDefaultDate: false,
		ServerHeader:       "gateway",

		ShutdownTimeout: 10 * time.Second,
	}
}
