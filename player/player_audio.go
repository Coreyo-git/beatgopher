//go:build cgo

package player

import(
	"encoding/binary"
	"bufio"
	"io"
	"time"
	"log"
	"layeh.com/gopus"

	"github.com/coreyo-git/beatgopher/services"
)

// Streams the audio to the voice channel.
func stream(p *Player) {
	// Ensure processes are killed when stream exits for any reason
	defer p.cleanupCurrentStream()

	vc := p.OnGetVoiceConnection()
	if vc == nil || !p.OnCheckVoiceConnection() {
		log.Println("Voice connection is invalid or disconnected, aborting stream")
		return
	}

	if vc.Status != 3 {
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
		// read full frame (EOF/error will be returned when stream is closed during cleanup)
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
			return
		case <-p.skip:
			log.Println("Song skipped by user")
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