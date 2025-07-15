package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"net/url"
	"github.com/coreyo-git/beatgopher/src/services"
)

func playHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Received your request for `%s`!", query),
		},
	})

	if err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong while trying to respond.",
		})
		log.Printf("Error responding to interaction: %v", err)
	}

	if isValidURL(query) {
		handleURL(s, i, query)
	} else {
		handleSearch(s, i, query)
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
func handleURL(s *discordgo.Session, i *discordgo.InteractionCreate, query string) {
	_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Getting song from: `%s`", query),
	})

	if err != nil {
		log.Panicf("Error during handleURL: %v", err)
	}
}

// called when the user's query is a song name
func handleSearch(s *discordgo.Session, i *discordgo.InteractionCreate, query string) {
	results, err := services.SearchYoutube(query)
	if err != nil {
		log.Panicf("Error during handleSearch: %v", err)
	}
	fmt.Println(results)
	
	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Searching for: `%s`", query),
	})
}

// checks if a string is a valid URL.
func isValidURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	return err == nil
}
