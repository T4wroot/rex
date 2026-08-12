package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the rex-node configuration
type Config struct {
	Token         string `yaml:"token"`
	Port          int    `yaml:"port"`
	TCPPort       int    `yaml:"tcp_port"`
	TLS           bool   `yaml:"tls"`
	CertFile      string `yaml:"cert_file"`
	KeyFile       string `yaml:"key_file"`
	AllowlistPath string `yaml:"allowlist"`
	LogLevel      string `yaml:"log_level"`
}

// DefaultConfig returns sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Port:          7443,
		TCPPort:       7444,
		TLS:           false,
		AllowlistPath: "/etc/rex/allowlist.yaml",
		LogLevel:      "info",
	}
}

// LoadConfig reads config from a YAML file
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file %q: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("cannot parse config file: %w", err)
	}

	if cfg.Token == "" {
		return nil, fmt.Errorf("token must be set in config")
	}

	log.Printf("[rex] config loaded from %s (port=%d, tls=%v)", path, cfg.Port, cfg.TLS)
	return cfg, nil
}
