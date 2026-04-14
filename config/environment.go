package config

import "fmt"

type Environment string

const (
	EnvLocal Environment = "local"
	EnvDev   Environment = "dev"
	EnvUAT   Environment = "uat"
	EnvProd  Environment = "prod"
)

func ParseEnvironment(raw string) (Environment, error) {
	env := Environment(raw)

	switch env {
	case EnvLocal, EnvDev, EnvUAT, EnvProd:
		return env, nil
	default:
		return "", fmt.Errorf("unknown environment: %s", raw)
	}
}

func (e Environment) ConfigPath() string {
	return fmt.Sprintf("configs/%s.toml", e)
}
