package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ActiveContext string             `yaml:"active_context" mapstructure:"active_context"`
	Contexts      map[string]Context `yaml:"contexts" mapstructure:"contexts"`
}

type Context struct {
	ApiUrl             string `yaml:"api_url" mapstructure:"api_url"`
	ApiKey             string `yaml:"api_key" mapstructure:"api_key"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify" mapstructure:"insecure_skip_verify"`
}

func SaveConfig(cfg *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(home, ".sd.yaml")

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

func LoadConfig() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Ensure maps are initialized
	if cfg.Contexts == nil {
		cfg.Contexts = make(map[string]Context)
	}

	// Default context if none exists
	if cfg.ActiveContext == "" {
		cfg.ActiveContext = "default"
	}

	// Set a default production URL if context is empty
	if _, ok := cfg.Contexts["default"]; !ok {
		cfg.Contexts["default"] = Context{
			ApiUrl: "https://api.saddledata.io",
		}
	}

	return &cfg, nil
}

func GetActiveContext(cfg *Config, override string) (string, Context, error) {
	name := cfg.ActiveContext
	if override != "" {
		name = override
	}

	ctx, ok := cfg.Contexts[name]
	if !ok {
		return name, Context{}, fmt.Errorf("context '%s' not found", name)
	}

	return name, ctx, nil
}
