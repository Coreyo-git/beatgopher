package services

import (
	"fmt"
	"os/exec"
	"strings"
)

type YoutubeResult struct {
	ID string `json:"id"`
	Channel string `json:"channel"`
	Title string `json:"title"`
	Duration string `json:"duration"`
	URL string `json:"url"`
	Thumbnail string `json:"thumbnail"`
}

func SearchYoutube(query string) (YoutubeResult, error) {
	result := YoutubeResult{}

	// yt-dlp args with custom output
	args := []string{
		"ytsearch:" + query,
		"--print", "%(id)s|%(channel)s|%(title)s|%(duration_string)s|%(webpage_url)s|%(thumbnail)s",
		"--skip-download",
	}
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

	line := strings.TrimSpace(string(output))
	if line == "" {
		return result, fmt.Errorf("no youtube result found for query: %s", query)
	}

	parts := strings.SplitN(line, "|", 6)
	if len(parts) < 6 {
		return result, fmt.Errorf("unexpected line format: %s", line)
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
