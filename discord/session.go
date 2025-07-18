package discord

import (
"fmt"
"github.com/bwmarrin/discordgo"
)

// Session provides helper methods for interacting with the Discord API.
type Session struct {
   *discordgo.Session
}
 
 // NewSession creates a new Session wrapper.
func NewSession(s *discordgo.Session) *Session {
	return &Session{s}
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
 
// JoinVoiceChannel finds the voice channel of the user who triggered the interaction and joins it.
func (s *Session) JoinVoiceChannel(i *discordgo.InteractionCreate) (*discordgo.VoiceConnection, error) {
	g, err := s.State.Guild(i.GuildID)
	if err != nil {
		return nil, fmt.Errorf("could not find guild: %w", err)
	}

	// Find the user's voice state.
	vs := findUserVoiceState(g, i.Member.User.ID)
	if vs == nil {
		return nil, fmt.Errorf("you are not in a voice channel")
	}

	// Join the user's voice channel.
	vc, err := s.ChannelVoiceJoin(i.GuildID, vs.ChannelID, false, true)
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