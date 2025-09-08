package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

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

	s.mu.Lock()
	s.VoiceConnection = vc
	s.mu.Unlock()

	return nil
}

func (s *Session) LeaveVoiceChannel() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.VoiceConnection != nil {
		s.VoiceConnection.Disconnect()
		s.VoiceConnection = nil
	}
}

// GetVoiceConnection returns the voice connection
func (s *Session) GetVoiceConnection() *discordgo.VoiceConnection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.VoiceConnection
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

// IsVoiceConnected checks if the bot is still connected to a voice channel
func (s *Session) IsVoiceConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.VoiceConnection == nil {
		return false
	}

	// Check if the connection is ready
	return s.VoiceConnection.Ready
}