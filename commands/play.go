package commands

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/discord"
	"github.com/coreyo-git/beatgopher/services"
)

func playHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	session := discord.GetOrCreateSession(s, i)

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
	err := session.InteractionRespond(i.Interaction, fmt.Sprintf("Received your request for `%s`!", query))

	if err != nil {
		session.InteractionRespond(i.Interaction, "Something went wrong while trying to respond.")
		log.Printf("Error responding to interaction: %v", err)
	}

	resultCh := make(chan services.YoutubeResult, 1)
	errCh := make (chan error, 1)

	log.Printf("Received song request for: %v", query)
	go func() {
		// Handles the search and gets the piped out audio stream
		song, err := handleSearch(session, i, query)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- song
	}()

	select {
	case song := <-resultCh:
		log.Printf("Adding song: %v", song.Title)
		session.Player.AddSong(i, &song)
	case err := <-errCh:
		log.Printf("Search Error: %v", err)
		session.FollowupMessage(i.Interaction, "Sorry I couldn't find that song or process the URL.")
	case <- time.After(30 * time.Second):
		log.Printf("Search timeout for query: %s", query)
		session.FollowupMessage(i.Interaction, "Search timed out. Please try again.")
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
