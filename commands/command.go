package commands

import "github.com/bwmarrin/discordgo"

// A map of all the registered commands, keyed by command name.
var Commands = make(map[string]Command)

// Command holds the definition and handler for a slash commands.
type Command struct {
	// Data sent to Discord to register the command.
	Definition *discordgo.ApplicationCommand
	// Function that runs when the command is used.
	Handler func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

type CommandRequest struct {
    CommandType string // "play" or "playlist"
    Interaction *discordgo.InteractionCreate
    Session     *discordgo.Session
    Handler     func(*discordgo.Session, *discordgo.InteractionCreate)
}