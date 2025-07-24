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
	TextChannelID string
	VoiceConnection *discordgo.VoiceConnection
}
 
 // NewSession creates a new Session wrapper.
func NewSession(s *discordgo.Session, guildID string, textChannelID string) *Session {
	return &Session{
		Session: s,
		GuildID: guildID,
		TextChannelID: textChannelID,
		VoiceConnection: nil,
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
	_, err := s.Session.ChannelMessageSend(s.TextChannelID, message)
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

	_, err := s.Session.ChannelMessageSendEmbed(s.TextChannelID, embed)
	if err != nil {
		return fmt.Errorf("error sending song embed: %v", err)
	}
	return nil
}

func (s *Session) SendQueueEmbed(songs []*services.YoutubeResult, currentPage int, totalPages int) error {
	embed := &discordgo.MessageEmbed{
		Title: "Queue",
		Color: 0x1DB954, // Spotify green, or choose any hex color
	}

	var description string
	if len(songs) == 0 {
		description = "The queue is empty."
	} else {
		for i, song := range songs {
			description += fmt.Sprintf("%d. [%s](%s) `[%s]`\n", i+1, song.Title, song.URL, song.Duration)
		}
	}

	embed.Description = description
	if totalPages > 1 {
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Page %d of %d", currentPage, totalPages),
		}
	}

	_, err := s.Session.ChannelMessageSendEmbed(s.TextChannelID, embed)
	if err != nil {
		return fmt.Errorf("error sending song embed: %v", err)
	}
	return nil
}
 
// JoinVoiceChannel finds the voice channel of the user who triggered the interaction and joins it.

func (s *Session) JoinVoiceChannel(i *discordgo.InteractionCreate) error {
	g, err := s.Session.State.Guild(i.GuildID)
	if err != nil {
		return fmt.Errorf("could not find guild: %w", err)
	}

	// Find the user's voice state.
	vs := findUserVoiceState(g, i.Member.User.ID)
	if vs == nil {
		return fmt.Errorf("you are not in a voice channel")
	}

	// Join the user's voice channel.
	vc, err := s.Session.ChannelVoiceJoin(s.GuildID, vs.ChannelID, false, true)
	if err != nil {
		s.FollowupMessage(i.Interaction, "Error joining voice channel")
		return fmt.Errorf("could not join voice channel: %w", err)
	}
	s.VoiceConnection = vc
	return nil
}

func (s *Session) LeaveVoiceChannel() {
	if s.VoiceConnection != nil {
		s.VoiceConnection.Disconnect()
		s.VoiceConnection = nil
	}
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