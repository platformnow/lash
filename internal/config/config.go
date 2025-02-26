package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	// General settings
	LogLevel    string `mapstructure:"log_level"`
	Interactive bool   `mapstructure:"interactive"`
	CacheTTL    int    `mapstructure:"cache_ttl"` // in minutes

	// GitHub settings
	GithubToken string `mapstructure:"github_token"`
	CatalogRepo struct {
		Owner string `mapstructure:"owner"`
		Name  string `mapstructure:"name"`
		Path  string `mapstructure:"path"`
	} `mapstructure:"catalog_repo"`

	// Kubernetes settings
	KubeConfig  string `mapstructure:"kubeconfig"`
	KubeContext string `mapstructure:"context"`
	DefaultNS   string `mapstructure:"namespace"`

	// Crossplane settings
	CrossplaneChart struct {
		Repository string                 `mapstructure:"repository"`
		Name       string                 `mapstructure:"name"`
		Version    string                 `mapstructure:"version"`
		Values     map[string]interface{} `mapstructure:"values"`
	} `mapstructure:"crossplane_chart"`

	// Default selections
	DefaultProviders []string `mapstructure:"default_providers"`
	DefaultPackages  []string `mapstructure:"default_packages"`
}

func LoadConfig() (*Config, error) {
	config := &Config{}

	// Set default locations
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath(filepath.Join(homeDir, ".lash"))

	// Set defaults
	viper.SetDefault("log_level", "info")
	viper.SetDefault("interactive", true)
	viper.SetDefault("cache_ttl", 60)
	viper.SetDefault("namespace", "crossplane-system")

	// Environment variables
	viper.SetEnvPrefix("LASH")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
