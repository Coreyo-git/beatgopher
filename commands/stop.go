package commands

import (
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/discord"
)

var notInChannelReplies = []string{
	"I'm not even in a voice channel! What am I supposed to stop? My existential crisis? 😅",
	"Umm... I'm not even playing anything? Are you trying to stop my thoughts? 🤔",
	"Stop what? I'm not even there! Did you forget to invite me to the party? 🎉",
	"You can't stop what was never started! *insert galaxy brain meme* 🧠",
	"Stop? I'm literally not even in the voice channel! Am I a ghost? 👻",
	"Error 404: Bot not found in voice channel. Please try again after summoning me! 🔮",
	"Bro, I'm not even there! Did you check if I'm actually in the voice channel? 😂",
	"I'm not playing anything! Are you trying to stop the silence? 🤫",
	"Stop what exactly? I'm not even connected! Did I miss the memo? 📝",
	"You want me to stop... nothing? That's already stopped! 🛑",
}

func stopHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	session := discord.GetOrCreateSession(s, i)

	if !session.IsVoiceConnected() {
		// Pick a random funny reply
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		message := notInChannelReplies[rng.Intn(len(notInChannelReplies))]
		session.InteractionRespond(i.Interaction, message)
		return
	}

	session.Player.Stop()

	session.InteractionRespond(i.Interaction, "Thanks for listening! 🎶")
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
