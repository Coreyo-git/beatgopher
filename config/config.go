package config

import (
	"log"
	"os"
)

// Config holds all configuration for the application.
type Config struct {
	Token string `json:"token"`
}

// Cfg is a global/package-level variable that holds the loaded configuration.
var Cfg *Config

// Init is executed when the package is imported.
func init() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
	}

	// Cfg is initialized with the loaded configuration values.
	Cfg = &Config{
		Token: token,
	}
}
