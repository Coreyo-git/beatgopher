package commands

import (
	"fmt"
	"log"
	"net/url"

	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/discord"
	"github.com/coreyo-git/beatgopher/player"
	"github.com/coreyo-git/beatgopher/services"
)

func playHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ds := discord.NewSession(s, i.GuildID, i.ChannelID)
	p := player.GetOrCreatePlayer(ds)

	p.Lock()
	defer p.Unlock()

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

	// Handles the search and gets the piped out audio stream
	song, err := handleSearch(ds, i, query)

	if err != nil {
		return
	}

	p.AddSong(i, &song)
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

// called when the user's query is a song name
func handleSearch(ds *discord.Session, i *discordgo.InteractionCreate, query string) (services.YoutubeResult, error) {
	youtubeService := &services.YoutubeService{}

	if isValidURL(query) {
		result, err := youtubeService.GetYoutubeInfo(query)
		if (err) != nil {
			return services.YoutubeResult{}, err
		}
		return result, nil
	}

	result, err := youtubeService.SearchYoutube(query)

	if err != nil {
		log.Printf("Error handling search: %v", err)
		ds.FollowupMessage(i.Interaction, "Sorry, I couldn't find that song or process the URL.")
		return services.YoutubeResult{}, err
	}

	return result, nil
}

// checks if a string is a valid URL.
func isValidURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	return err == nil
}
