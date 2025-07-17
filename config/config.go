package config

import (
	"log"
	"os"
	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	Token 	string `json:"token"`
}

// Cfg is a global/package-level variable that holds the loaded configuration.
var Cfg *Config

// Init is executed when the package is imported.
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file.", err)
	}

	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("TOKEN environment variable not set.")
	}

    // Cfg is initialized with the loaded configuration values.
	Cfg = &Config{
		Token: token,
	}
}