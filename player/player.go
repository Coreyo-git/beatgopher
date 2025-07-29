package player

import (
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
	Queue     *queue.Queue
	Session   *discord.Session
	CurrentStream *services.AudioStream
	IsPlaying bool
	stop      chan bool
	mu        sync.Mutex
}

var (
	playersMutex sync.Mutex
	// Map of guild IDs to players.
	players = make(map[string]*Player)
)

func NewPlayer(ds *discord.Session) *Player {
	return &Player{
		Queue:     queue.NewQueue(),
		Session:   ds,
		CurrentStream: nil,
		IsPlaying: false,
		stop:      make(chan bool),
		mu:        sync.Mutex{},
	}
}

// Gets or creates the player for a guild
func GetOrCreatePlayer(ds *discord.Session) *Player {
	playersMutex.Lock()
	defer playersMutex.Unlock()
	player, exists := players[ds.GuildID]
	if !exists {
		player = NewPlayer(ds)
		players[ds.GuildID] = player
	}

	return player
}

// Adds a song to the queue and starts playback if the player is not already playing.
func (player *Player) AddSong(i *discordgo.InteractionCreate, song *services.YoutubeResult) {
	player.Queue.Enqueue(song)

	player.mu.Lock()
	isAlreadyPlaying := player.IsPlaying
	player.mu.Unlock()

	if !isAlreadyPlaying {
		player.mu.Lock()
		player.IsPlaying = true
		player.mu.Unlock()

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
		p.CurrentStream.Stdout = stdout
		if err != nil {
			log.Printf("Error in setupAudioOutput: %v", err)
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

// Streams the audio to the voice channel.
func stream(i *discordgo.InteractionCreate, p *Player) {
	// Join the voice channel of the user who sent the command.
	err := p.Session.JoinVoiceChannel(i)

	if err != nil {
		return
	}

	if !p.Session.VoiceConnection.Ready {
		time.Sleep(1 * time.Millisecond)
	}

	p.Session.VoiceConnection.Speaking(true)

	defer p.Session.VoiceConnection.Speaking(false)
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

	// Reads raw PCM data from the stream
	pcm := make([]int16, frameSize*channels)
	for {
		// read full frame
		err := binary.Read(p.CurrentStream.Stdout, binary.LittleEndian, &pcm)
		if err != nil {
			// if stream end io.eof will be returned
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				log.Println("Stream Finished.")
				return // end
			}
			log.Printf("Error reading from ffmpeg stream: %v", err)
			return
		}

		// Encode the PCM data into an Opus packet.
		opus, err := encoder.Encode(pcm, frameSize, maxBytes)
		if err != nil {
			log.Printf("Error encoding audio to opus: %v", err)
			return
		}
		select {
		case p.Session.VoiceConnection.OpusSend <- opus:
		case <-p.stop:
			return
		case <-time.After(2 * time.Second):
			log.Println("Timeout sending opus packet, disconnecting.")
			return
		}
	}
}

// Sets up audio output from a YouTube result.
func setupAudioOutput(result *services.YoutubeResult) (io.ReadCloser, error) {
	// Consumer/Producer pipe to buffer to stream
	pipeReader, pipeWriter := io.Pipe()

	// go routine producing data from ffmpeg and filling the pipe
	go func() {
		defer pipeWriter.Close()

		CurrentStream, err := services.NewAudioStream(result.URL)
		if err != nil {
			pipeWriter.CloseWithError(err)
			return
		}

		// copy data from ffmpeg to output pipe
		// should block until song is finished or error
		_, err = io.Copy(pipeWriter, CurrentStream.FfmpegStdout)
		if err != nil {
			log.Printf("Error Copying audio stream: %v", err)
		}

		if err := CurrentStream.Ytdlp.Wait(); err != nil {
			log.Printf("Error waiting for ytdlp: %v", err)
		}

		if err := CurrentStream.Ffmpeg.Wait(); err != nil {
			log.Printf("Error waiting for ffmpeg: %v", err)
		}
	}()

	return pipeReader, nil
}


