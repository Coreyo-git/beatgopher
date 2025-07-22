package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/services"
)

// Session provides helper methods for interacting with the Discord API.
type Session struct {
    Session *discordgo.Session
   	GuildID string
	ChannelID string
	VoiceChannelID string
}
 
 // NewSession creates a new Session wrapper.
func NewSession(s *discordgo.Session, guildID string, channelID string) *Session {
	return &Session{
		Session: s,
		GuildID: guildID,
		ChannelID: channelID,
		VoiceChannelID: "",
	}
}
 
// InteractionRespond is a wrapper for s.InteractionRespond that simplifies sending a basic message.
func (s *Session) InteractionRespond(i *discordgo.Interaction, content string) error {
	return s.Session.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
        Content: content,
		},
	})
}
 
// FollowupMessage is a wrapper for s.FollowupMessageCreate that simplifies sending a followup message.
func (s *Session) FollowupMessage(i *discordgo.Interaction, content string) error {
	_, err := s.Session.FollowupMessageCreate(i,true, &discordgo.WebhookParams{
		Content: content,
	})
	return err
}

func (s *Session) SendChannelMessage(message string) error {
	_, err := s.Session.ChannelMessageSend(s.ChannelID, message)
	if err != nil {
		return fmt.Errorf("error sending channel message: %v", err)
	}
	return nil
}

func (s *Session) SendSongEmbed(song *services.YoutubeResult, footer string) error {
	embed := &discordgo.MessageEmbed{
		Title:       song.Title,
		URL:         song.URL,
		Description: fmt.Sprintf("Channel: **%s**\nDuration: `%s`", song.Channel, song.Duration),
		Color:       0x1DB954, // Spotify green, or choose any hex color
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: song.Thumbnail,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: footer,
		},
	}

	_, err := s.Session.ChannelMessageSendEmbed(s.ChannelID, embed)
	if err != nil {
		return fmt.Errorf("error sending song embed: %v", err)
	}
	return nil
}
 
// JoinVoiceChannel finds the voice channel of the user who triggered the interaction and joins it.
func (s *Session) JoinVoiceChannel(i *discordgo.InteractionCreate) (*discordgo.VoiceConnection, error) {
	g, err := s.Session.State.Guild(i.GuildID)
	if err != nil {
		return nil, fmt.Errorf("could not find guild: %w", err)
	}

	// Find the user's voice state.
	vs := findUserVoiceState(g, i.Member.User.ID)
	if vs == nil {
		return nil, fmt.Errorf("you are not in a voice channel")
	}

	// Join the user's voice channel.
	vc, err := s.Session.ChannelVoiceJoin(s.GuildID, vs.ChannelID, false, true)
	if err != nil {
		s.FollowupMessage(i.Interaction, "Error joining voice channel")
		return nil, fmt.Errorf("could not join voice channel: %w", err)
	}

	return vc, nil
}

// Finds the user's voice state in the guild.
func findUserVoiceState(guild *discordgo.Guild, userID string) *discordgo.VoiceState {
	for _, vs := range guild.VoiceStates {
		if vs.UserID == userID {
			return vs
		}
	}
	return nil
}