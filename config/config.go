package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server ServerConfig `toml:"server"`
	OAuth  OAuthConfig  `toml:"oauth"`
}

type ServerConfig struct {
	Port string `toml:"port"`
}

type OAuthConfig struct {
	AccountURL  string `toml:"account_url"`
	ClientID    string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
	RedirectURI string `toml:"redirect_uri"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	// defaults
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8081"
	}
	return &cfg, nil
}
