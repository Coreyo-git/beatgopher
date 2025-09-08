package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/discord"
)

func stopHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	session := discord.GetOrCreateSession(s, i)

	session.Player.Stop()

	session.InteractionRespond(i.Interaction, "Stopped and disconnected.")
}

func init() {
	Commands["stop"] = Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "stop",
			Description: "Stops and disconnects the bot.",
		},
		Handler: stopHandler,
	}
}



