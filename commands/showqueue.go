package commands

import (
	"log"
	"math"

	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/discord"
)

// songsPerPage is the number of songs to display on each page of the queue.
const songsPerPage = 10

func showqueueHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Create a new Discord session wrapper.
	session := discord.GetOrCreateSession(s, i)

	// Get the songs from the queue.
	songs := session.Queue.GetSongs()
	if len(songs) == 0 {
		session.InteractionRespond(i.Interaction, "The queue is empty.")
		return
	}

	// Get the page number from the command options.
	page := 1
	if i.ApplicationCommandData().Options != nil {
		page = int(i.ApplicationCommandData().Options[0].IntValue())
	}

	// Calculate the total number of pages.
	totalPages := int(math.Ceil(float64(len(songs)) / float64(songsPerPage)))
	if page > totalPages {
		page = totalPages
	}

	// Calculate the start and end indices for the current page.
	start := (page - 1) * songsPerPage
	end := start + songsPerPage
	if end > len(songs) {
		end = len(songs)
	}

	// Respond to the interaction to prevent time out.
	err := session.InteractionRespond(i.Interaction, "Loading queue...")
	if err != nil {
		log.Printf("Error responding to interaction: %v", err)
		return
	}

	err = session.SendQueueEmbed(songs[start:end], page, totalPages)
	if err != nil {
		session.FollowupMessage(i.Interaction, "Something went wrong while trying to show the queue.")
		log.Printf("Error sending queue embed: %v", err)
	}
}

func init() {
	Commands["showqueue"] = Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "showqueue",
			Description: "Shows the current song queue.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "page",
					Description: "The page of the queue to view.",
					Required:    false,
				},
			},
		},
		Handler: showqueueHandler,
	}
}
