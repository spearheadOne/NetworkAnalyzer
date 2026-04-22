package config

import (
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Topology   TopologyConfig   `toml:"topology"`
	OpenSearch OpenSearchConfig `toml:"opensearch"`
	Collector  CollectorConfig  `toml:"collector"`
}

type TopologyConfig struct {
	Addr string `toml:"Address"`
}

type CollectorConfig struct {
	WorkersNum int `toml:"workers-num"`
	QueueSize  int `toml:"queue-size"`
}

type OpenSearchConfig struct {
	Host         string `toml:"host"`
	CounterIndex string `toml:"counter-index"`
	FlowIndex    string `toml:"flow-index"`
}

func Load(envFlag string) (*Config, error) {
	env, err := ParseEnvironment(envFlag)
	if err != nil {
		log.Fatal(err)
	}

	path := env.ConfigPath()

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("decode toml config %q: %w", path, err)
	}

	return &cfg, nil
}
