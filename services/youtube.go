package services

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// YoutubeResult holds the structured data for a single YouTube video,
// parsed from the output of the yt-dlp command.
type YoutubeResult struct {
	ID        string `json:"id"`
	Channel   string `json:"channel"`
	Title     string `json:"title"`
	Duration  string `json:"duration_string"`
	URL       string `json:"webpage_url"`
	Thumbnail string `json:"thumbnail"`
}

// GetYoutubeInfo fetches metadata for a single YouTube video URL by calling yt-dlp.
func GetYoutubeInfo(url string) (YoutubeResult, error) {
	result := YoutubeResult{}

	// Create a new slice with the
	args := buildYtdlpArgs(url)

	cmd := exec.Command("yt-dlp", args...)

	// Capture stdout and stderr
	output, err := cmd.Output()
	if err != nil {
		// Print stderr for debugging
		if ee, ok := err.(*exec.ExitError); ok {
			fmt.Println("yt-dlp error output:", string(ee.Stderr))
		}
		fmt.Println("Command error:", err)
		return result, err
	}

	return parseYoutubeOutput(output)
}

// SearchYoutube performs a search on YouTube using yt-dlp's "ytsearch:" prefix
// and returns the first video result.
func SearchYoutube(query string) (YoutubeResult, error) {
	result := YoutubeResult{}

	// yt-dlp args with custom output
	args := buildYtdlpArgs("ytsearch:" + query)

	cmd := exec.Command("yt-dlp", args...)
	output, err := cmd.Output()

	if err != nil {
		// Print stderr for debugging
		if ee, ok := err.(*exec.ExitError); ok {
			fmt.Println("yt-dlp error output:", string(ee.Stderr))
		}
		fmt.Println("Command error:", err)
		return result, err
	}

	return parseYoutubeOutput(output)
}

// GetYoutubePlaylistInfo retrieves metadata for multiple videos from a YouTube playlist URL.
// It can limit the number of videos processed and optionally randomize the playlist order.
func GetYoutubePlaylistInfo(playlistURL string, total int64, randomizeSongs bool) ([]YoutubeResult, error) {
	results := []YoutubeResult{}
	// yt-dlp args with custom output
	args := []string{
		"--print", "%(id)s|%(channel)s|%(title)s|%(duration_string)s|%(webpage_url)s|%(thumbnail)s",
		"--flat-playlist",
		"--skip-download",
		playlistURL,
	}
	// The --playlist-random flag tells yt-dlp to shuffle the playlist before processing.
	if randomizeSongs {
		args = append(args, "--playlist-random")
	}

	cmd := exec.Command("yt-dlp", args...)
	output, err := cmd.Output()

	if err != nil {
		// Print stderr for debugging
		if ee, ok := err.(*exec.ExitError); ok {
			fmt.Println("yt-dlp playlist error output:", string(ee.Stderr))
		}
		fmt.Println("Command error:", err)
		return results, err
	}

	// yt-dlp returns one line of output for each video in the playlist.
	lines := strings.Split(string(output), "\n")
	var i int64 = 1
	// Process each line, stopping if the desired total is reached.
	for _, line := range lines {
		if i > total {
			break
		}
		if line == "" {
			continue
		}
		result, err := parseYoutubeOutput([]byte(line))
		if err != nil {
			log.Printf("Failed to parse line: %v", line)
		}
		results = append(results, result)
		i++
	}

	return results, nil
}

// parseYoutubeOutput takes the raw byte output from a yt-dlp command
// and parses it into a YoutubeResult struct.
// It expects a single line of text with fields delimited by "|".
func parseYoutubeOutput(output []byte) (YoutubeResult, error) {
	result := YoutubeResult{}

	line := strings.TrimSpace(string(output))
	if line == "" {
		return result, fmt.Errorf("empty output from yt-dlp")
	}

	// The output format is defined by the --print argument in the yt-dlp command.
	// Example: "VIDEO_ID|CHANNEL_NAME|VIDEO_TITLE|DURATION|VIDEO_URL|THUMBNAIL_URL"
	parts := strings.SplitN(line, "|", 6)
	if len(parts) < 6 {
		return result, fmt.Errorf("unexpected output format: %s", line)
	}

	result = YoutubeResult{
		ID:        parts[0],
		Channel:   parts[1],
		Title:     parts[2],
		Duration:  parts[3],
		URL:       parts[4],
		Thumbnail: parts[5],
	}

	return result, nil
}

// buildYtdlpArgs constructs the command-line arguments for yt-dlp for fetching
// metadata for a single video.
// It takes a 'target', which can be a video URL or a search query string.
func buildYtdlpArgs(target string) []string {
	return []string{
		target,
		// --skip-download: Prevents downloading the video file.
		"--skip-download",
		// --no-playlist: If a video URL is part of a playlist, only process the video.
		"--no-playlist",
		// --flat-playlist: Do not extract video information from playlist pages, just list entries.
		"--flat-playlist",
		"--no-warnings",
		"--no-check-certificate",
		// --geo-bypass: Attempt to bypass geographic restrictions.
		"--geo-bypass",
		// --print: Defines a custom output format. We use "|" as a separator.
		"--print", "%(id)s|%(channel)s|%(title)s|%(duration_string)s|%(webpage_url)s|%(thumbnail)s",
	}
}
