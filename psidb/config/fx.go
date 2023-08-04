package config

import "go.uber.org/fx"

var Module = fx.Module(
	"config",

	fx.Provide(NewConfig),
	fx.Provide(NewLocalResourceManager),
)

func NewConfig() (*Config, error) {
	cfg, err := ReadConfig()

	if err != nil {
		return nil, err
	}

	if err := cfg.InitializeDefaults(); err != nil {
		return nil, err
	}

	return cfg, nil
}
