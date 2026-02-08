package config

import (
	"asum/pkg/db"
	"asum/pkg/engine"
	"asum/pkg/mailer"
	"asum/pkg/maxmind"
	"asum/pkg/rdb"
	"asum/pkg/token"
	"fmt"

	"github.com/spf13/viper"
)

type AppConfig struct {
	BaseURL  string         `mapstructure:"baseURL" yaml:"baseURL"`
	Engine   engine.Config  `mapstructure:"engine" yaml:"engine"`
	Email    mailer.Config  `mapstructure:"mail" yaml:"mail"`
	MaxMind  maxmind.Config `mapstructure:"maxmind" yaml:"maxmind"`
	JWT      token.Config   `mapstructure:"jwt" yaml:"jwt"`
	Redis    rdb.Config     `mapstructure:"redis" yaml:"redis"`
	Postgres db.Config      `mapstructure:"postgres" yaml:"postgres"`
}

func Load(path string) (AppConfig, error) {
	cfg := AppConfig{}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}
