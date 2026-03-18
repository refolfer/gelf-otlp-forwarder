package internal

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"net"
	"os"
)

type Config struct {
	InboundPort  int    `yaml:"inbound_port"`
	OutboundHost string `yaml:"outbound_host"`
	OutboundPort int    `yaml:"outbound_port"`
}

// LoadConfig reads the configuration from config.yaml and validates it
func LoadConfig() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		log.Println("Env CONFIG_PATH not found. Use default path: /config/config.yaml")
		configPath = "/config/config.yaml"
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	config := &Config{}
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// validateConfig checks if the configuration is valid
func validateConfig(cfg *Config) error {
	// Validate inbound port
	if cfg.InboundPort < 1 || cfg.InboundPort > 65535 {
		return fmt.Errorf("invalid inbound_port: %d (must be 1-65535)", cfg.InboundPort)
	}

	// Validate outbound port
	if cfg.OutboundPort < 1 || cfg.OutboundPort > 65535 {
		return fmt.Errorf("invalid outbound_port: %d (must be 1-65535)", cfg.OutboundPort)
	}

	// Validate outbound host
	if cfg.OutboundHost == "" {
		return fmt.Errorf("outbound_host cannot be empty")
	}

	// Try to resolve the hostname/IP
	if net.ParseIP(cfg.OutboundHost) == nil {
		// Not an IP, check if it resolves as hostname
		_, err := net.LookupHost(cfg.OutboundHost)
		if err != nil {
			log.Printf("Warning: outbound_host '%s' may not be reachable: %v\n", cfg.OutboundHost, err)
		}
	}

	return nil
}
