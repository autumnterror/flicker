package config

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/autumnterror/breezynotes/pkg/utils/format"
	"github.com/spf13/viper"
)

type Config struct {
	Uri                  string
	TokenKey             string
	AccessTokenLifeTime  time.Duration
	RefreshTokenLifeTime time.Duration
	Port                 int
}

// MustSetup return config and panic if error
func MustSetup() *Config {
	cfg, err := setup()
	if err != nil {
		log.Panic(err)
	}
	return cfg
}

// setup create config structure
func setup() (*Config, error) {
	const op = "config.setup"
	configPath := flag.String("config", "./build/deploy/configs/local.yaml", "path to config file")
	flag.Parse()
	viper.SetConfigFile(*configPath)

	var cfg struct {
		Db                   string
		Pw                   string
		User                 string
		DataSource           string        `mapstructure:"data_source"`
		PortPostgres         int           `mapstructure:"port_postgres"`
		TokenKey             string        `mapstructure:"token_key"`
		AccessTokenLifeTime  time.Duration `mapstructure:"access_token_life"`
		RefreshTokenLifeTime time.Duration `mapstructure:"refresh_token_life"`
		Port                 int
		Mode                 string
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, format.Error(op, err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, format.Error(op, err)
	}

	if cfg.Mode == "DEV" {
		log.Println(format.Struct(cfg), fmt.Sprintf("URI: postgres://%s:%s@%s:%d/%s?sslmode=disable",
			cfg.User, cfg.Pw, cfg.DataSource, cfg.PortPostgres, cfg.Db))
	}

	return &Config{
		Uri: fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			cfg.User, cfg.Pw, cfg.DataSource, cfg.PortPostgres, cfg.Db),
		TokenKey:             cfg.TokenKey,
		AccessTokenLifeTime:  cfg.AccessTokenLifeTime,
		RefreshTokenLifeTime: cfg.RefreshTokenLifeTime,
		Port:                 cfg.Port,
	}, nil
}
