package player

import (
	"encoding/binary"
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

type Player struct {
	Queue *queue.Queue
	Session *discord.Session
	IsPlaying bool
	GuildID string
	ChannelID string
	mu sync.Mutex
}

var (
	playersMutex sync.Mutex 
	players = make(map[string]*Player) 
)

 // NewSession creates a new Session wrapper.
 func NewPlayer(guildID string, ds *discord.Session, channelID string) *Player {
	return &Player{
		Queue: queue.NewQueue(),
		Session: ds,
		IsPlaying: false,
		GuildID: guildID,
		ChannelID: channelID,
		mu: sync.Mutex{},
	}
}

func GetOrCreatePlayer(guildID string, ds *discord.Session, channelID string) *Player {
	playersMutex.Lock()
	defer playersMutex.Unlock()
	player, exists := players[guildID]
	if !exists {
		player = NewPlayer(guildID, ds, channelID)
		players[guildID] = player
	} 
	
	return player
}

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
		player.Session.SendSongEmbed(player.ChannelID, song, "Queued to play.")
	}
} 

func (p *Player) handlePlaybackLoop(i *discordgo.InteractionCreate) {
	for {
		song := p.Queue.Dequeue()
		if song == nil {
			break
		}
		p.Session.SendSongEmbed(p.ChannelID, song, "Playing!")

		stdout, err := setupAudioOutput(song)
		if err != nil {
			log.Printf("Error in setupAudioOutput: %v", err)
		}

		stream(i, stdout, p)
	}
	p.mu.Lock()
	p.IsPlaying = false
	p.mu.Unlock()
}

func stream(i *discordgo.InteractionCreate, stream io.ReadCloser, p *Player) {
	// Join the voice channel of the user who sent the command.
	vc, err := p.Session.JoinVoiceChannel(i)

	if err != nil {
		return
	}

	if !vc.Ready {
		time.Sleep(1* time.Millisecond)
	}

	vc.Speaking(true)

	defer vc.Speaking(false)
	const (
		channels int = 2
		frameRate int = 48000
		frameSize int = 960
		maxBytes int = 1275
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
		err := binary.Read(stream, binary.LittleEndian, &pcm)
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
		case vc.OpusSend <- opus:
		case <- time.After(2* time.Second):
			log.Println("Timeout sending opus packet, disconnecting.")
			return
		}
	}
}

// Called when the user's query is a URL.
func setupAudioOutput(result *services.YoutubeResult) (io.ReadCloser, error) {
	// Consumer/Producer pipe to buffer to stream
	pipeReader, pipeWriter := io.Pipe()

	// go routine producing data from ffmpeg and filling the pipe
	go func() {
		defer pipeWriter.Close()

		stdout, err := services.GetAudioStream(result.URL)
		if err != nil {
			pipeWriter.CloseWithError(err)
			return
		}

		// copy data from ffmpeg to output pipe
		// should block until song is finished or error
		_, err = io.Copy(pipeWriter, stdout)
		if err != nil {
			log.Printf("Error Copying audio stream: %v", err)
		}
	}()

	return pipeReader, nil
}


