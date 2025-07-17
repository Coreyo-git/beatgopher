package commands

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/url"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/discord"
	"github.com/coreyo-git/beatgopher/services"
	"layeh.com/gopus"
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
	vc, err := ds.JoinVoiceChannel(i)

	if err != nil {
		ds.FollowupMessage(i.Interaction, "Error joining voice channel")
		log.Printf("Error joining voice channel: %v", err)
		return
	}

	// Handles the search and gets the piped out audio stream
	stdoutStream, err := handleSearch(ds, i, query)

	if err != nil {
		return
	}

	go playStream(vc, stdoutStream)
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
func setupAudioOutput(ds *discord.Session, i *discordgo.InteractionCreate, result services.YoutubeResult) (io.ReadCloser, error) {
	ds.FollowupMessage(i.Interaction, fmt.Sprintf("Getting song from: `%s`", result.Title))

	stdout, err := services.GetAudioStream(result.URL)
	if err != nil {
		return nil, err
	}

	return stdout, err
}

// called when the user's query is a song name
func handleSearch(ds *discord.Session, i *discordgo.InteractionCreate, query string) (io.ReadCloser, error) {
	if isValidURL(query){
		result, err := services.GetYoutubeInfo(query)
		if(err) != nil {
			return nil, err
		}
		return setupAudioOutput(ds, i, result)
	}

	result, err := services.SearchYoutube(query)

	if err != nil {
		return nil, err
	}

	return setupAudioOutput(ds, i, result)
}

// checks if a string is a valid URL.
func isValidURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	return err == nil
}
