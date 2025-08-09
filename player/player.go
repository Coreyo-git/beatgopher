package player

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/discord"
	"github.com/coreyo-git/beatgopher/queue"
	"github.com/coreyo-git/beatgopher/services"
	"layeh.com/gopus"
)

// Player represents a music player for a single guild.
type Player struct {
	Queue         queue.QueueInterface
	Session       discord.DiscordSessionInterface
	CurrentStream *services.AudioStream
	IsPlaying     bool
	stop          chan bool
	mu            sync.Mutex
}

var (
	playersMutex sync.Mutex
	// Map of guild IDs to players.
	players = make(map[string]*Player)
)

func NewPlayer(ds discord.DiscordSessionInterface) *Player {
	return &Player{
		Queue:         queue.NewQueue(),
		Session:       ds,
		CurrentStream: nil,
		IsPlaying:     false,
		stop:          make(chan bool),
		mu:            sync.Mutex{},
	}
}

// Gets or creates the player for a guild
func GetOrCreatePlayer(ds discord.DiscordSessionInterface) *Player {
	playersMutex.Lock()
	defer playersMutex.Unlock()
	player, exists := players[ds.GetGuildID()]
	if !exists {
		player = NewPlayer(ds)
		players[ds.GetGuildID()] = player
	}

	return player
}

// Adds a song to the queue and starts playback if the player is not already playing.
func (player *Player) AddSong(i *discordgo.InteractionCreate, song *services.YoutubeResult) {
	player.Queue.Enqueue(song)

	player.mu.Lock()
	defer player.mu.Unlock()

	if !player.IsPlaying {
		player.IsPlaying = true
		go player.handlePlaybackLoop(i)
	} else {
		player.Session.SendSongEmbed(song, "Queued to play.")
	}
}

func (player *Player) AddSongs(i *discordgo.InteractionCreate, songs []services.YoutubeResult) {
	for j := 0; j < len(songs); j++ {
		if j == 0 {
			fmt.Printf("Playing song: %v\n", &songs[j])
			player.AddSong(i, &songs[j])
			continue
		}
		fmt.Printf("Adding song to queue: %v\n", &songs[j])
		player.Queue.Enqueue(&songs[j])
	}
}

// handlePlaybackLoop is the main loop for playing songs from the queue.
func (p *Player) handlePlaybackLoop(i *discordgo.InteractionCreate) {
	for {
		song := p.Queue.Dequeue()
		if song == nil {
			break
		}
		p.Session.SendSongEmbed(song, "Playing!")

		stdout, err := setupAudioOutput(song)
		if err != nil {
			log.Printf("Error in setupAudioOutput: %v", err)
		}

		// Create a new AudioStream struct for this song
		p.CurrentStream = &services.AudioStream{
			Stdout: stdout,
		}

		stream(i, p)
	}

	p.Stop()
}

// Skip current song.
func (p *Player) Skip() bool {
	if p.IsPlaying {
		p.stop <- true
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

	// Leave the voice channel
	p.Session.LeaveVoiceChannel()
}

// IsPlayerPlaying returns true if the player is currently playing
func (p *Player) IsPlayerPlaying() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.IsPlaying
}

// GetQueue returns the queue interface
func (p *Player) GetQueue() queue.QueueInterface {
	return p.Queue
}

// GetSession returns the discord session interface
func (p *Player) GetSession() discord.DiscordSessionInterface {
	return p.Session
}

// Streams the audio to the voice channel.
func stream(i *discordgo.InteractionCreate, p *Player) {
	// Join the voice channel of the user who sent the command.
	err := p.Session.JoinVoiceChannel(i)
	if err != nil {
		log.Printf("Failed to join voice channel: %v", err)
		return
	}

	vc := p.Session.GetVoiceConnection()
	if !vc.Ready {
		log.Println("Voice connection not ready, waiting...")
		time.Sleep(100 * time.Millisecond) // Increased wait time
	}

	vc.Speaking(true)
	defer vc.Speaking(false)

	const (
		channels  int = 2
		frameRate int = 48000
		frameSize int = 960
		maxBytes  int = 1275
	)

	encoder, err := gopus.NewEncoder(frameRate, channels, gopus.Audio)
	if err != nil {
		log.Printf("Error creating Opus encoder: %v", err)
		return
	}

	// Debugging counters
	var (
		framesProcessed int64
		timeouts        int64
		errors          int64
		startTime       = time.Now()
	)

	// Log stats every 5 seconds
	statsTicker := time.NewTicker(5 * time.Second)
	defer statsTicker.Stop()

	go func() {
		for range statsTicker.C {
			elapsed := time.Since(startTime)
			log.Printf("Audio Stats - Frames: %d, Timeouts: %d, Errors: %d, Duration: %v",
				framesProcessed, timeouts, errors, elapsed)
		}
	}()

	// Reads raw PCM data from the stream
	pcm := make([]int16, frameSize*channels)
	for {
		// read full frame
		err := binary.Read(p.CurrentStream.Stdout, binary.LittleEndian, &pcm)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				log.Printf("Stream finished after %d frames", framesProcessed)
				return
			}
			log.Printf("Error reading from audio stream: %v", err)
			errors++
			return
		}

		// Encode the PCM data into an Opus packet.
		opus, err := encoder.Encode(pcm, frameSize, maxBytes)
		if err != nil {
			log.Printf("Error encoding audio to opus: %v", err)
			errors++
			return
		}

		select {
		case vc.OpusSend <- opus:
			framesProcessed++
		case <-p.stop:
			log.Println("Playback stopped by user")
			return
		case <-time.After(5 * time.Second):
			log.Printf("Timeout sending opus packet (frame %d)", framesProcessed)
			timeouts++
			// Don't return, try to recover
			continue
		}
	}
}

// Sets up audio output from a YouTube result.
func setupAudioOutput(result *services.YoutubeResult) (io.ReadCloser, error) {
	// Consumer/Producer pipe to buffer to stream
	pipeReader, pipeWriter := io.Pipe()

	go func() {
		defer pipeWriter.Close()

		log.Printf("Starting audio stream for: %s", result.Title)
		CurrentStream, err := services.NewAudioStream(result.URL)
		if err != nil {
			log.Printf("Error creating audio stream: %v", err)
			pipeWriter.CloseWithError(err)
			return
		}

		// Add buffering to smooth out the stream
		bufferedReader := bufio.NewReaderSize(CurrentStream.FfmpegStdout, 64*1024) // 64KB buffer

		// Copy with progress logging
		written, err := io.Copy(pipeWriter, bufferedReader)
		if err != nil {
			log.Printf("Error copying audio stream after %d bytes: %v", written, err)
		} else {
			log.Printf("Audio stream completed, %d bytes processed", written)
		}

		// Wait for processes to complete
		if err := CurrentStream.Ytdlp.Wait(); err != nil {
			log.Printf("yt-dlp process error: %v", err)
		}

		if err := CurrentStream.Ffmpeg.Wait(); err != nil {
			log.Printf("ffmpeg process error: %v", err)
		}

		log.Println("Audio stream cleanup completed")
	}()

	return pipeReader, nil
}
