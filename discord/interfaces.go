package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/services"
)

// DiscordSessionInterface defines the contract for Discord session operations
type DiscordSessionInterface interface {
	// InteractionRespond sends a response to an interaction
	InteractionRespond(i *discordgo.Interaction, content string) error

	// FollowupMessage sends a followup message to an interaction
	FollowupMessage(i *discordgo.Interaction, content string) error

	// SendChannelMessage sends a message to the text channel
	SendChannelMessage(message string) error

	// SendSongEmbed sends an embed message for a song
	SendSongEmbed(song *services.YoutubeResult, footer string) error

	// SendQueueEmbed sends an embed message for the queue
	SendQueueEmbed(songs []*services.YoutubeResult, currentPage int, totalPages int) error

	// JoinVoiceChannel joins the voice channel of the user who triggered the interaction
	JoinVoiceChannel(i *discordgo.InteractionCreate) error

	// LeaveVoiceChannel leaves the current voice channel
	LeaveVoiceChannel()

	// GetGuildID returns the guild ID
	GetGuildID() string

	// GetTextChannelID returns the text channel ID
	GetTextChannelID() string

	// GetVoiceConnection returns the voice connection
	GetVoiceConnection() *discordgo.VoiceConnection

	// IsVoiceConnected checks if the bot is still connected to a voice channel
	IsVoiceConnected() bool
}

// Verify that Session implements DiscordSessionInterface at compile time
var _ DiscordSessionInterface = (*Session)(nil)
