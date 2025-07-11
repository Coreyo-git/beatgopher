package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/coreyo-git/beatgopher/config"

	"github.com/bwmarrin/discordgo"
)

func main() {
	// Create a new Discord session using the provided bot token.
	session, err := discordgo.New("Bot " + config.Cfg.Token)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	// Add a handler for the messageCreate event.
	// This function will be called every time a new message is created on any channel that the authenticated user has access to.
	// TODO: Give it a configurable channel scope or maybe just give it permission for one music channel??
	session.AddHandler(messageCreate)

	// messages and voice states.
	// https://discord.com/developers/docs/topics/gateway#gateway-intents
	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates

	// Open a websocket connection to Discord.
	err = session.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	session.Close()
}

// messageCreate will be called every time a new message is created.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Bast test setup response
	if m.Content == "hello" {
		s.ChannelMessageSend(m.ChannelID, "world!")
	}
}
