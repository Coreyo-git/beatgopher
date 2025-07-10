package main

import (
	"log"
	"os"
	"fmt"

	// "github.com/bmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("TOKEN")

	fmt.Println("Token:", token)
	// sess, err := discordgo.New("Bot " + token)
}
