package services

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func (as *AudioStream) Close() {
	if as.Ytdlp != nil && as.Ytdlp.Process != nil {
		as.Ytdlp.Process.Kill()
	}
	if as.Ffmpeg != nil && as.Ffmpeg.Process != nil {
		as.Ffmpeg.Process.Kill()
	}
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
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting working directory: %v", err)
		return
	}

	// Read directory contents
	files, err := os.ReadDir(wd)
	if err != nil {
		log.Printf("Error reading directory: %v", err)
		return
	}

	// Look for fragment files (--Frag followed by numbers)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		filename := file.Name()
		// Check if file matches fragment pattern (--Frag followed by digits)
		if strings.HasPrefix(filename, "--Frag") && len(filename) > 6 {
			// Check if the rest is numeric
			suffix := filename[6:] // Remove "--Frag" prefix
			isNumeric := true
			for _, char := range suffix {
				if char < '0' || char > '9' {
					isNumeric = false
					break
				}
			}
			
			if isNumeric {
				fullPath := filepath.Join(wd, filename)
				if err := os.Remove(fullPath); err != nil {
					log.Printf("Error removing fragment file %s: %v", filename, err)
				} else {
					log.Printf("Removed fragment file: %s", filename)
				}
			}
		}
	}
}
