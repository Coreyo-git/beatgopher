package player

import (
	"encoding/binary"
	"io"
	"log"
	"sync"
	"time"
	"fmt"

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
	mu sync.Mutex
}

var players = make(map[string]*Player)

 // NewSession creates a new Session wrapper.
 func NewPlayer(guildID string, ds *discord.Session) *Player {
	return &Player{
		Queue: queue.NewQueue(),
		Session: ds,
		IsPlaying: false,
		GuildID: guildID,
		mu: sync.Mutex{},
	}
}

func GetOrCreatePlayer(guildID string, ds *discord.Session) *Player {
	player, exists := players[guildID]
	if !exists {
		player = NewPlayer(guildID, ds)
		players[guildID] = player
	} 
	
	return player
}

func (*Player) AddSong(i *discordgo.InteractionCreate, player *Player, song *services.YoutubeResult) {
	player.Queue.Enqueue(song)

	player.mu.Lock()
	isAlreadyPlaying := player.IsPlaying
	player.mu.Unlock()
	
	if !isAlreadyPlaying {
		player.mu.Lock()
		player.IsPlaying = true
		player.mu.Unlock()

		go handlePlaybackLoop(i, player.GuildID)
	} else {
		player.Session.FollowupMessage(i.Interaction, fmt.Sprintf("Added song to queue: %v", song.Title))
	}
} 

func handlePlaybackLoop(i *discordgo.InteractionCreate, guild string) {
	player := GetOrCreatePlayer(guild, nil)

	for {
		song := player.Queue.Dequeue()
		if song == nil {
			break
		}
		player.Session.FollowupMessage(i.Interaction, fmt.Sprintf("Playing song: %v", song.Title))

		stdout, err := setupAudioOutput(song)
		if err != nil {
			log.Printf("Error in setupAudioOutput: %v", err)
		}

		Stream(i, stdout, player)
	}
	player.mu.Lock()
	player.IsPlaying = false
	player.mu.Unlock()
}

func Stream(i *discordgo.InteractionCreate, stream io.ReadCloser, p *Player) {
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

	stdout, err := services.GetAudioStream(result.URL)
	if err != nil {
		return nil, err
	}

	return stdout, err
}


