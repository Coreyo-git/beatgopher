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



func SearchYoutube(query string) ([]YoutubeResult, error) {
	results := []YoutubeResult{}

	// yt-dlp args with custom output
	args := []string{
		"ytsearch3:" + query,
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
		return results, err
	}

	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 6)
		if len(parts) < 6 {
			fmt.Println("unexpected line format:", line)
			continue
		}

		result := YoutubeResult{
			ID:        parts[0],
			Channel:   parts[1],
			Title:     parts[2],
			Duration:  parts[3],
			URL:       parts[4],
			Thumbnail: parts[5],
		}

		results = append(results, result)
	}

	fmt.Println(string(output))
	
	return results, nil
}
