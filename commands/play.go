package commands

import (
	"fmt"
	"log"
	"net/url"

	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/discord"
	"github.com/coreyo-git/beatgopher/services"
)

func playHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ds := discord.NewSession(s)
	var query string

	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options

	// convert the slice into a map
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	query = optionMap["query"].StringValue()


	// Acknowledge command and reply to avoid timeout.
	err := ds.InteractionRespond(i.Interaction, fmt.Sprintf("Received your request for `%s`!", query))

	if err != nil {
		ds.InteractionRespond(i.Interaction, "Something went wrong while trying to respond.")
		log.Printf("Error responding to interaction: %v", err)
	}

	// Join the voice channel of the user who sent the command.
	_, err = ds.JoinVoiceChannel(i)
	if err != nil {
		ds.FollowupMessage(i.Interaction, "Error joining voice channel")
		log.Printf("Error joining voice channel: %v", err)
		return
	}

	if isValidURL(query) {
		handleURL(ds, i, query)
	} else {
		handleSearch(ds, i, query)
	}
}

func init() {
	minLength := 3
	Commands["play"] = Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "play",
			Description: "Plays a song from a URL",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "query",
					Description: "The URL of the song or a search term.",
					Required:    true,
					MinLength:   &minLength,
				},
			},
		},
		Handler: playHandler,
	}
}

// Called when the user's query is a URL.
func handleURL(ds *discord.Session, i *discordgo.InteractionCreate, query string) {
	err := ds.FollowupMessage(i.Interaction, fmt.Sprintf("Getting song from: `%s`", query))

	if err != nil {
		log.Panicf("Error during handleURL: %v", err)
	}
}

// called when the user's query is a song name
func handleSearch(ds *discord.Session, i *discordgo.InteractionCreate, query string) {
	result, err := services.SearchYoutube(query)
	if err != nil {
		log.Panicf("Error during handleSearch: %v", err)
	}

	// TODO: do something with the result

	ds.FollowupMessage(i.Interaction, fmt.Sprintf("Found song: `%s`", result.URL))
}

// checks if a string is a valid URL.
func isValidURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	return err == nil
}
