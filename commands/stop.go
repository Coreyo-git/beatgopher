package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/discord"
	"github.com/coreyo-git/beatgopher/player"
)

func stopHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ds := discord.NewSession(s, i.GuildID, i.ChannelID)
	p := player.GetOrCreatePlayer(ds)

	p.Stop()

	ds.InteractionRespond(i.Interaction, "Stopped and disconnected.")
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



