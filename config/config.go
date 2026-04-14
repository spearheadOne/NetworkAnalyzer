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

	if cfg.OpenSearch.Host == "" {
		return nil, fmt.Errorf("opensearch.host must not be empty in %q", path)
	}
	if cfg.OpenSearch.CounterIndex == "" {
		return nil, fmt.Errorf("opensearch.counter-index must not be empty in %q", path)
	}
	if cfg.OpenSearch.FlowIndex == "" {
		return nil, fmt.Errorf("opensearch.flow-index must not be empty in %q", path)
	}

	return &cfg, nil
}
