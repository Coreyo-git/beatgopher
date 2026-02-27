package player

import (
	"fmt"
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/queue"
	"github.com/coreyo-git/beatgopher/services"
)

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
}

// Player represents a music player for a single guild.
type Player struct {
	CurrentStream *services.AudioStream
	Queue         queue.QueueInterface
	IsPlaying     bool
	stop          chan bool
	skip          chan bool
	mu            sync.RWMutex

	OnSendEmbedMessage     func(song *services.YoutubeResult, content string) error
	OnCheckVoiceConnection func() bool
	OnGetVoiceConnection   func() *discordgo.VoiceConnection
	OnLeaveVoiceChannel	   func() 
}

func NewPlayer(
	queue queue.QueueInterface,
	onSendEmbedMessage func(song *services.YoutubeResult, content string) error,
	onCheckVoiceConnection func() bool,
	onGetVoiceConnection func() *discordgo.VoiceConnection,
	onLeaveVoiceChannel func(),
) *Player {
	return &Player{
		CurrentStream: nil,
		Queue:         queue,
		IsPlaying:     false,
		stop:          make(chan bool),
		skip:          make(chan bool),
		mu:            sync.RWMutex{},

		OnSendEmbedMessage:     onSendEmbedMessage,
		OnCheckVoiceConnection: onCheckVoiceConnection,
		OnGetVoiceConnection:   onGetVoiceConnection,
		OnLeaveVoiceChannel: 	onLeaveVoiceChannel,
	}
}

// Adds a song to the queue and starts playback if the player is not already playing.
func (p *Player) AddSong(i *discordgo.InteractionCreate, song *services.YoutubeResult) {
	p.Queue.Enqueue(song)
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.IsPlaying {
		log.Printf("Starting playback loop")
		p.IsPlaying = true
		go p.playbackLoop()
	} else {
		p.OnSendEmbedMessage(song, "Added to queue!")
	}
}

func (p *Player) AddSongs(i *discordgo.InteractionCreate, songs []services.YoutubeResult) {
	for j := 0; j < len(songs); j++ {
		if j == 0 {
			p.AddSong(i, &songs[j])
			continue
		}
		fmt.Printf("Adding song to queue: %v", &songs[j])
		p.Queue.Enqueue(&songs[j])
	}
}

// playbackLoop is the main loop for playing songs from the queue.
// It runs in its own goroutine.
func (p *Player) playbackLoop() {
	for {
		select{
		case <-p.stop:
			return
		default:
			song := p.Queue.Dequeue()
			if song == nil {
				p.Stop();
				return
			}
	
			p.OnSendEmbedMessage(song, "Playing!")
	
			_, err := setupAudioOutput(song, p)
			if err != nil {
				log.Printf("Error in setupAudioOutput: %v", err)
				continue // Skip this song and move to the next one
			}
	
			log.Printf("Starting stream.")
			stream(p)
		}
	}
}

// Skip current song.
func (p *Player) Skip() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.IsPlaying {
		// Non-blocking send to skip channel
		select {
		case p.skip <- true:
		default:
		}
		return true
	}
	return false
}

// Stops the player and clears the queue.
func (p *Player) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Clear the queue
	p.Queue = queue.NewQueue()
	p.IsPlaying = false

	// Non-blocking send to stop channel - if nothing is receiving,
	// the playback loop has already exited, so we just move on
	select {
	case p.stop <- true:
	default:
	}

	p.OnLeaveVoiceChannel()
}

// IsPlayerPlaying returns true if the player is currently playing
func (p *Player) IsPlayerPlaying() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.IsPlaying
}

// cleanupCurrentStream kills the yt-dlp and ffmpeg processes for the current stream
func (p *Player) cleanupCurrentStream() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.CurrentStream != nil {
		p.CurrentStream.Close()
		p.CurrentStream = nil
	}
}

// GetQueue returns the queue interface
func (p *Player) GetQueue() queue.QueueInterface {
	return p.Queue
}


