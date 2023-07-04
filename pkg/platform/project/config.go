package project

type Config struct {
	Project struct {
		Name             string   `toml:"name"`
		EnabledLanguages []string `toml:"enabled_languages"`
	}

	Modules []ModuleConfig `toml:"module"`
}

type ModuleConfig struct {
	Name     string `toml:"name"`
	Path     string `toml:"path"`
	Language string `toml:"language"`
}
