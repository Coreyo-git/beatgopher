package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/coreyo-git/beatgopher/commands"
	"github.com/coreyo-git/beatgopher/config"
	"github.com/coreyo-git/beatgopher/player"

	"github.com/bwmarrin/discordgo"
)

func main() {
	// Create a new Discord session using the provided bot token.
	session, err := discordgo.New("Bot " + config.Cfg.Token)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	// Add a handler for interactions e.g.. /play
	session.AddHandler(interactionCreate)

	// Add a handler for voice state changes to detect disconnections
	session.AddHandler(voiceStateUpdate)

	// messages and voice states.
	// https://discord.com/developers/docs/topics/gateway#gateway-intents
	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates

	// This function will be called once the bot is connected and ready.
	// It will also call registerCommands function
	session.AddHandler(onReady)

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

// interactionCreate will be called every time a new interaction is created.
func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Check interaction type
	if i.Type == discordgo.InteractionApplicationCommand {
		// Look for command with matching name in registry
		if cmd, ok := commands.Commands[i.ApplicationCommandData().Name]; ok {
			// if exists call relative handler
			cmd.Handler(s, i)
		}
	}
}

func onReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Println("Registering commands...")
	registerCommands(s)
}

// voiceStateUpdate handles voice state changes to detect when the bot gets disconnected
func voiceStateUpdate(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	// Check if this is our bot's voice state
	if vsu.UserID != s.State.User.ID {
		return
	}

	// If the bot left a voice channel (ChannelID is empty), clean up the player
	if vsu.ChannelID == "" {
		log.Printf("Bot was disconnected from voice channel in guild: %s", vsu.GuildID)
		player.HandleBotDisconnection(vsu.GuildID)
	}
}

// Iterates over command registry adding each command
func registerCommands(s *discordgo.Session) {
	for _, cmd := range commands.Commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd.Definition)
		if err != nil {
			log.Fatalf("Cannot create slash command '%s': %v", cmd.Definition.Name, err)
		}
	}
}
