package services

import (
	"fmt"
	"io"
	"log"
	"os/exec"
)

type AudioStream struct {
	FfmpegStdout io.ReadCloser
	Ytdlp        *exec.Cmd
	Ffmpeg       *exec.Cmd
	Stdout       io.ReadCloser
}

func NewAudioStream(url string) (*AudioStream, error) {
	ffmpegStdout, ytdlp, ffmpeg, err := setupAudioStream(url)
	if err != nil {
		return nil, err
	}

	return &AudioStream{
		FfmpegStdout: ffmpegStdout,
		Ytdlp:        ytdlp,
		Ffmpeg:       ffmpeg,
		Stdout:       nil,
	}, nil
}

// GetAudioStream returns a reader with the raw audio data from a YouTube URL.
func setupAudioStream(url string) (io.ReadCloser, *exec.Cmd, *exec.Cmd, error) {
	ytdlpArgs := []string{
		url,
		"-f", "bestaudio",
		"-o", "-", // output to stdout
	}
	ytdlp := exec.Command("yt-dlp", ytdlpArgs...)

	ffmpegArgs := []string{
		"-i", "pipe:0", // input from stdin
		"-f", "s16le",
		"-ar", "48000",
		"-ac", "2",
		"pipe:1", // output to stdout
	}
	ffmpeg := exec.Command("ffmpeg", ffmpegArgs...)

	// Pipe yt-dlp's stdout to ffmpeg's stdin
	ytdlpStdout, err := ytdlp.StdoutPipe()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error creating yt-dlp stdout pipe: %w", err)
	}
	ffmpeg.Stdin = ytdlpStdout

	// Get ffmpeg's stdout pipe
	ffmpegStdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error creating ffmpeg stdout pipe: %w", err)
	}

	// Start both processes
	if err := ytdlp.Start(); err != nil {
		return nil, nil, nil, fmt.Errorf("error starting yt-dlp: %w", err)
	}

	if err := ffmpeg.Start(); err != nil {
		return nil, nil, nil, fmt.Errorf("error starting ffmpeg: %w", err)
	}

	return ffmpegStdout, ytdlp, ffmpeg, nil
}

// Cleanup forcefully terminates the FFmpeg and yt-dlp processes and closes streams
func (as *AudioStream) Cleanup() {
	if as == nil {
		return
	}

	// Close the stdout reader first to signal downstream consumers
	if as.FfmpegStdout != nil {
		as.FfmpegStdout.Close()
	}
	if as.Stdout != nil {
		as.Stdout.Close()
	}

	// Terminate FFmpeg process
	if as.Ffmpeg != nil && as.Ffmpeg.Process != nil {
		log.Printf("Terminating FFmpeg process (PID: %d)", as.Ffmpeg.Process.Pid)
		if err := as.Ffmpeg.Process.Kill(); err != nil {
			log.Printf("Error killing FFmpeg process: %v", err)
		}
		// Wait for process to actually terminate (with timeout)
		go func() {
			as.Ffmpeg.Wait()
			log.Printf("FFmpeg process terminated")
		}()
	}

	// Terminate yt-dlp process
	if as.Ytdlp != nil && as.Ytdlp.Process != nil {
		log.Printf("Terminating yt-dlp process (PID: %d)", as.Ytdlp.Process.Pid)
		if err := as.Ytdlp.Process.Kill(); err != nil {
			log.Printf("Error killing yt-dlp process: %v", err)
		}
		// Wait for process to actually terminate (with timeout)
		go func() {
			as.Ytdlp.Wait()
			log.Printf("yt-dlp process terminated")
		}()
	}
}
