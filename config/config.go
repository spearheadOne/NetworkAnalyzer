package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	OpenSearch OpenSearchConfig `toml:"opensearch"`
}

type OpenSearchConfig struct {
	Host         string `toml:"host"`
	CounterIndex string `toml:"counter-index"`
	FlowIndex    string `toml:"flow-index"`
}

func Load(env Environment) (*Config, error) {
	path := env.ConfigPath()

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("decode toml config %q: %w", path, err)
	}
	return &cfg, nil
}
