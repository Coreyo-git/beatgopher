package config

import (
	"os"
	"github.com/joho/godotenv"
	"fmt"
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
		fmt.Printf("No .env file found.", err)
	}

	token := os.Getenv("TOKEN")
	if token == "" {
		token = os.Getenv("DISCORD_TOKEN")
		fmt.Printf("ENV Token not set as 'TOKEN', using 'DISCORD_TOKEN' instead.")
	}

    // Cfg is initialized with the loaded configuration values.
	Cfg = &Config{
		Token: token,
	}
}