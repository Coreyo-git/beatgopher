package player

import (
	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/discord"
	"github.com/coreyo-git/beatgopher/queue"
	"github.com/coreyo-git/beatgopher/services"
)

// PlayerInterface defines the contract for player operations
type PlayerInterface interface {
	// AddSong adds a song to the queue and starts playback if not already playing
	AddSong(i *discordgo.InteractionCreate, song *services.YoutubeResult)

	// AddSongs adds multiple songs to the queue
	AddSongs(i *discordgo.InteractionCreate, songs []services.YoutubeResult)

	// Skip skips the current song
	Skip() bool

	// Stop stops the player and clears the queue
	Stop()

	// IsPlayerPlaying returns true if the player is currently playing
	IsPlayerPlaying() bool

	// GetQueue returns the queue interface
	GetQueue() queue.QueueInterface

	// GetSession returns the discord session interface
	GetSession() discord.DiscordSessionInterface
}

// PlayerManagerInterface defines the contract for managing players across guilds
type PlayerManagerInterface interface {
	// GetOrCreatePlayer gets or creates a player for a guild
	GetOrCreatePlayer(ds discord.DiscordSessionInterface) PlayerInterface
}

// Verify that Player implements PlayerInterface at compile time
var _ PlayerInterface = (*Player)(nil)
