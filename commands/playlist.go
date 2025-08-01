package commands

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/discord"
	"github.com/coreyo-git/beatgopher/player"
	"github.com/coreyo-git/beatgopher/services"
)

func playlistHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ds := discord.NewSession(s, i.GuildID, i.ChannelID)
	p := player.GetOrCreatePlayer(ds)

	var query string
	var total int64 = 25
	var random bool = false

	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options

	// convert the slice into a map
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	if(optionMap["total"] != nil) {
		total = optionMap["total"].IntValue()
	}

	if(optionMap["random"] != nil) {
		random = optionMap["random"].BoolValue()
	}

	query = optionMap["url"].StringValue()
	
	// Acknowledge command and reply to avoid timeout.
	err := ds.InteractionRespond(i.Interaction, fmt.Sprintf("Received your request for `%s`!", query))

	if err != nil {
		ds.InteractionRespond(i.Interaction, "Something went wrong while trying to respond.")
		log.Printf("Error responding to interaction: %v", err)
	}

	// Handles the search and gets the piped out audio stream
	songs, err := handlePlaylist(ds, i, query, total, random)

	if err != nil {
		return
	} 
	fmt.Println("Adding songs from playlist")
	p.AddSongs(i, songs)
}

func init() {
	Commands["playlist"] = Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "playlist",
			Description: "Fills the queue with songs from the playlist",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "url",
					Description: "The URL of the playlist.",
					Required:    true,
				},
				{
					Type:		discordgo.ApplicationCommandOptionInteger,
					Name: 		"total",
					Description: "Total amount of songs to play from the playlist (Default 25)",
				},
				{
					Type: 		discordgo.ApplicationCommandOptionBoolean,
					Name:		"random",
					Description:"Randomize the songs from the playlist",
				},
			},
		},
		Handler: playlistHandler,
	}
}

// called when the user's query is a song name
func handlePlaylist(d *discord.Session, i *discordgo.InteractionCreate, q string, total int64, random bool) ([]services.YoutubeResult, error) {
	if isValidPlaylistURL(q){
		results, err := services.GetYoutubePlaylistInfo(q, total, random)
		if(err) != nil {
			return []services.YoutubeResult{}, err
		}
		return results, nil
	}

	return []services.YoutubeResult{}, nil
}

// checks if a string is a valid playlist URL.
func isValidPlaylistURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	return err == nil && strings.Contains(s, "list=")
}


