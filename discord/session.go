package discord

import (
	"fmt"
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/player"
	"github.com/coreyo-git/beatgopher/queue"
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

	JoinIfVoiceIsNotConnected(i *discordgo.InteractionCreate) error

	// GetGuildID returns the guild ID
	GetGuildID() string

	// GetTextChannelID returns the text channel ID
	GetTextChannelID() string

	// GetVoiceConnection returns the voice connection
	GetVoiceConnection() *discordgo.VoiceConnection

	// IsVoiceConnected checks if the bot is still connected to a voice channel
	IsVoiceConnected() bool

	// function to remove from the queue 
	RemoveFromQueue(song *services.YoutubeResult) bool
}

// Session provides helper methods for interacting with the Discord API.
type Session struct {
	Session         *discordgo.Session
	GuildID         string
	TextChannelID   string
	VoiceConnection *discordgo.VoiceConnection
	Player          player.PlayerInterface
	Queue           queue.QueueInterface
	mu              sync.RWMutex
}

var (
	sessionsMutex sync.Mutex
	// Map of guild IDs to players.
	sessions = make(map[string]*Session)
)

// NewSession creates a new Session wrapper.
func newSession(s *discordgo.Session, i *discordgo.InteractionCreate) *Session {
	queue := queue.NewQueue()
	session := &Session{
		Session:         s,
		GuildID:         i.GuildID,
		TextChannelID:   i.ChannelID,
		VoiceConnection: nil,
		Queue:           queue,
		Player:			 nil,
		mu:              sync.RWMutex{},
	}
	session.mu.Lock()
	defer session.mu.Unlock()

	session.Player = player.NewPlayer(
		queue,
		session.SendSongEmbed,
		session.IsVoiceConnected,
		session.GetVoiceConnection,
		session.LeaveVoiceChannel,
	)

	return session
}

func GetOrCreateSession(s *discordgo.Session, i *discordgo.InteractionCreate) *Session {
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()

	session, exists := sessions[i.GuildID]
	if !exists {
		session = newSession(s, i)
		sessions[i.GuildID] = session
	}

	return session
}

// HandleBotDisconnection cleans up session state when the bot gets disconnected from a guild
func HandleBotDisconnection(guildID string) {
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()

	if session, exists := sessions[guildID]; exists {
		session.mu.Lock()
		defer session.mu.Unlock()
		log.Printf("Cleaning up player state for disconnected guild: %s", guildID)

		// Reset player state
		session.Queue.Clear()
		session.Player.Stop()

		// Clear voice connection reference
		if session.GetVoiceConnection() != nil {
			session.VoiceConnection = nil
		}

		// Remove the player from the map since it's no longer valid
		delete(sessions, guildID)

		log.Printf("Session cleanup completed for guild: %s", guildID)
	}
}

func (s *Session) RemoveFromQueue(song *services.YoutubeResult) bool {
	return s.Queue.RemoveFromQueue(song)
}

// GetGuildID returns the guild ID
func (s *Session) GetGuildID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.GuildID
}

// GetTextChannelID returns the text channel ID
func (s *Session) GetTextChannelID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.TextChannelID
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
	_, err := s.Session.FollowupMessageCreate(i, true, &discordgo.WebhookParams{
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
		Footer: &discordgo.MessageEmbedFooter{
			Text: footer,
		},
	}
	if song.Thumbnail != "NA" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: song.Thumbnail,
		}
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
