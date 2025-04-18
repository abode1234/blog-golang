package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server struct {
		Port        int    `toml:"port"`
		ContentDir  string `toml:"content_dir"`
		TemplateDir string `toml:"template_dir"`
	} `toml:"server"`

	Site struct {
		Name        string `toml:"name"`
		Author      string `toml:"author"`
		Description string `toml:"description"`
		Social struct {
			Github   string `toml:"github"`
			X        string `toml:"x"`
			Linkedin string `toml:"linkedin"`
		} `toml:"social"`
	} `toml:"site"`
}

func Load(path string) (*Config, error) {
	cfg := new(Config)
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config load failed: %w", err)
	}

	if _, err := toml.Decode(string(file), cfg); err != nil {
		return nil, fmt.Errorf("config parse failed: %w", err)
	}
	return cfg, nil
}
