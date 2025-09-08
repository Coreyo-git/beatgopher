package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/discord"
	"github.com/coreyo-git/beatgopher/services"
)

func removeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	session := discord.GetOrCreateSession(s, i)

	// Get the songs from the queue to check if it's empty
	songs := session.Queue.GetSongs()
	if len(songs) == 0 {
		session.InteractionRespond(i.Interaction, "The queue is empty. Nothing to remove.")
		return
	}

	// Get command options
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	var removedSong *string
	var err error

	// Check if user provided a position number
	if positionOpt, exists := optionMap["position"]; exists {
		position := int(positionOpt.IntValue())
		removedSong, err = removeByPosition(session, position, songs)
	} else if queryOpt, exists := optionMap["query"]; exists {
		// Remove by search query (title or partial title match)
		query := queryOpt.StringValue()
		removedSong, err = removeByQuery(session, query, songs)
	} else {
		session.InteractionRespond(i.Interaction, "Please provide either a position number or a search query to remove a song.")
		return
	}

	if err != nil {
		session.InteractionRespond(i.Interaction, err.Error())
		return
	}

	if removedSong != nil {
		session.InteractionRespond(i.Interaction, fmt.Sprintf("✅ Removed **%s** from the queue.", *removedSong))
	} else {
		session.InteractionRespond(i.Interaction, "❌ Could not find the specified song to remove.")
	}
}

// removeByPosition removes a song at the specified position (1-indexed)
func removeByPosition(s discord.DiscordSessionInterface, position int, songs []*services.YoutubeResult) (*string, error) {
	if position < 1 || position > len(songs) {
		return nil, fmt.Errorf("❌ Invalid position. Please specify a position between 1 and %d", len(songs))
	}

	// Convert to 0-indexed
	songToRemove := songs[position-1]

	if s.RemoveFromQueue(songToRemove) {
		return &songToRemove.Title, nil
	}

	return nil, nil
}

// removeByQuery removes the first song that matches the query (case-insensitive partial match)
func removeByQuery(s discord.DiscordSessionInterface, query string, songs []*services.YoutubeResult) (*string, error) {
	query = strings.ToLower(query)

	for _, song := range songs {
		if strings.Contains(strings.ToLower(song.Title), query) {
			if s.RemoveFromQueue(song) {
				return &song.Title, nil
			}
		}
	}

	return nil, fmt.Errorf("❌ No song found matching '%s'", query)
}

func init() {
	Commands["remove"] = Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "remove",
			Description: "Remove a song from the queue by position or search query.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "position",
					Description: "The position of the song to remove (1-indexed, use /showqueue to see positions).",
					Required:    false,
					MinValue:    func() *float64 { v := 1.0; return &v }(),
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "query",
					Description: "Search for a song by title to remove (partial matches allowed).",
					Required:    false,
					MinLength:   func() *int { v := 1; return &v }(),
				},
			},
		},
		Handler: removeHandler,
	}
}
