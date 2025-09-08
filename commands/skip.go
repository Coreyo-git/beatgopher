package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/discord"
)

func skipHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	session := discord.GetOrCreateSession(s, i)
	if session.Player.Skip() {
		session.InteractionRespond(i.Interaction, "Skipped the current song.")
	} else {
		session.InteractionRespond(i.Interaction, "Nothing to skip.")
	}
}

func init() {
	Commands["skip"] = Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "skip",
			Description: "Skips the current song.",
		},
		Handler: skipHandler,
	}
}
