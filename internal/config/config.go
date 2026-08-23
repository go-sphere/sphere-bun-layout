package config

import (
	"fmt"

	"github.com/go-sphere/confstore"
	"github.com/go-sphere/confstore/codec"
	"github.com/go-sphere/confstore/provider/file"
	"github.com/go-sphere/sphere-bun-layout/internal/pkg/database"
	"github.com/go-sphere/sphere-bun-layout/internal/server/api"
	"github.com/go-sphere/sphere-bun-layout/internal/server/docs"
	"github.com/go-sphere/sphere/log/zapx"
	"github.com/go-sphere/sphere/utils/secure"
)

var BuildVersion = "dev"

type Config struct {
	Environments map[string]string `json:"environments" yaml:"environments"`
	Log          zapx.Config       `json:"log" yaml:"log"`
	API          api.Config        `json:"api" yaml:"api"`
	Docs         docs.Config       `json:"docs" yaml:"docs"`
	Database     database.Config   `json:"database" yaml:"database"`
}

func NewEmptyConfig() *Config {
	return &Config{
		Environments: map[string]string{},
		Log: zapx.Config{
			File: zapx.FileConfig{
				FileName:   "./var/log/sphere.log",
				MaxSize:    10,
				MaxBackups: 10,
				MaxAge:     10,
			},
			Console: zapx.ConsoleConfig{},
			Level:   "info",
		},
		API: api.Config{
			JWT: secure.RandString(32),
			HTTP: api.HTTPConfig{
				Address: "0.0.0.0:8899",
			},
		},
		Docs: docs.Config{
			Address: "0.0.0.0:9999",
			Targets: docs.Targets{
				API: "http://localhost:8899",
			},
		},
		Database: database.Config{
			Location: "./var/db.sqlite3",
		},
	}
}

func NewConfig(path string) (*Config, error) {
	config, err := confstore.Load[Config](file.New(path), codec.JsonCodec())
	if err != nil {
		return nil, err
	}
	if config.Log.Level == "" {
		config.Log.Level = "info"
	}
	if config.API.JWT == "" {
		return nil, fmt.Errorf("api jwt must be non-empty")
	}
	return config, nil
}
