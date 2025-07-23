package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/discord"
	"github.com/coreyo-git/beatgopher/player"
)

func skipHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ds := discord.NewSession(s, i.GuildID, i.ChannelID)
	p := player.GetOrCreatePlayer(ds)

	if p.Skip() {
		ds.InteractionRespond(i.Interaction, "Skipped the current song.")
	} else {
		ds.InteractionRespond(i.Interaction, "Nothing to skip.")
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
