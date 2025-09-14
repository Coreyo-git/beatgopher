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
	"github.com/coreyo-git/beatgopher/queue"
	"github.com/coreyo-git/beatgopher/services"
	"layeh.com/gopus"
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
}

func NewPlayer(
	queue queue.QueueInterface,
	onSendEmbedMessage func(song *services.YoutubeResult, content string) error,
	onCheckVoiceConnection func() bool,
	onGetVoiceConnection func() *discordgo.VoiceConnection,
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
	}
}

// Start begins the player's playback loop as a goroutine.
// It should be called once per player.
func (p *Player) Start() {
	go p.playbackLoop()
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
		fmt.Printf("Adding song to queue: %v", &songs[j])
		p.Queue.Enqueue(&songs[j])
	}
}

// playbackLoop is the main loop for playing songs from the queue.
// It runs in its own goroutine.
func (p *Player) playbackLoop() {
	for {
		p.mu.RLock()
		// if !p.IsPlaying{
		// 	p.mu.RUnlock()
		// }
		select{
		case <-p.stop:
			p.mu.RUnlock()
			return
		default:
			p.mu.RUnlock()
			song := p.Queue.Dequeue()
			if song == nil {
				log.Printf("No more songs in queue.")
				p.mu.Lock()
				p.IsPlaying = false
				p.mu.Unlock()
				continue
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

	// Clean up current stream if it exists
	if p.CurrentStream != nil {
		p.CurrentStream.Cleanup()
		p.CurrentStream = nil
	}

	// Clear the queue
	p.Queue = queue.NewQueue()
	p.IsPlaying = false

	// Signal stop to stream() to interrupt playback
	select {
	case p.stop <- true:
	default:
	}
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

// Streams the audio to the voice channel.
func stream(p *Player) {
	vc := p.OnGetVoiceConnection()
	if vc == nil || !p.OnCheckVoiceConnection() {
		log.Println("Voice connection is invalid or disconnected, aborting stream")
		return
	}

	if !vc.Ready {
		log.Println("Voice connection not ready, waiting...")
		time.Sleep(100 * time.Millisecond) // Increased wait time

		// Check again after waiting
		if !p.OnCheckVoiceConnection() {
			log.Println("Voice connection lost while waiting, aborting stream")
			return
		}
	}

	vc.Speaking(true)
	defer vc.Speaking(false)

	// Ensure cleanup happens when stream function exits
	defer func() {
		if p.CurrentStream != nil {
			p.CurrentStream.Cleanup()
		}
	}()

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
			// Periodically check if we're still connected (every 100 frames)
			if framesProcessed%100 == 0 && !p.OnCheckVoiceConnection() {
				log.Println("Voice connection lost during streaming, stopping playback")
				return
			}
		case <-p.stop:
			log.Println("Playback stopped by user")
			// Clean up the current stream
			if p.CurrentStream != nil {
				p.CurrentStream.Cleanup()
			}
			return
		case <-p.skip:
			log.Println("Song skipped by user")
			if p.CurrentStream != nil {
				p.CurrentStream.Cleanup()
			}
			return
		case <-time.After(5 * time.Second):
			log.Printf("Timeout sending opus packet (frame %d)", framesProcessed)
			timeouts++
			// Check if connection is still valid on timeout
			if !p.OnCheckVoiceConnection() {
				log.Println("Voice connection lost during timeout, stopping playback")
				return
			}
			// Don't return, try to recover
			continue
		}
	}
}

// Sets up audio output from a YouTube result.
func setupAudioOutput(result *services.YoutubeResult, p *Player) (io.ReadCloser, error) {
	// Consumer/Producer pipe to buffer to stream
	pipeReader, pipeWriter := io.Pipe()

	log.Printf("Starting audio stream for: %s", result.Title)
	CurrentStream, err := services.NewAudioStream(result.URL)
	if err != nil {
		log.Printf("Error creating audio stream: %v", err)
		pipeWriter.CloseWithError(err)
		return nil, err
	}

	// Set the CurrentStream on the player for cleanup purposes before starting the copy goroutine
	p.mu.Lock()
	p.CurrentStream = CurrentStream
	p.CurrentStream.Stdout = pipeReader
	p.mu.Unlock()

	go func() {
		defer pipeWriter.Close()

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
