package services

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type YoutubeResult struct {
	ID string `json:"id"`
	Channel string `json:"channel"`
	Title string `json:"title"`
	Duration string `json:"duration_string"`
	URL string `json:"webpage_url"`
	Thumbnail string `json:"thumbnail"`
}

func GetYoutubeInfo(url string) (YoutubeResult, error) {
	result := YoutubeResult{}

	args := []string{
		url,
		"--skip-download",
		"--no-playlist", 
		"--flat-playlist",
		"--no-warnings",
		"--no-check-certificate",
		"--geo-bypass" ,
		"--print", "%(id)s|%(channel)s|%(title)s|%(duration_string)s|%(webpage_url)s|%(thumbnail)s",
	}
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

func SearchYoutube(query string) (YoutubeResult, error) {
	result := YoutubeResult{}

	// yt-dlp args with custom output
	args := []string{
		"ytsearch:" + query,
		"--skip-download",
		"--no-playlist", 
		"--flat-playlist",
		"--no-warnings",
		"--no-check-certificate",
		"--geo-bypass" ,
		"--print", "%(id)s|%(channel)s|%(title)s|%(duration_string)s|%(webpage_url)s|%(thumbnail)s",
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

	return parseYoutubeOutput(output)
}

func GetYoutubePlaylistInfo(playlistURL string, total int64, randomizeSongs bool) ([]YoutubeResult, error) {
	results := []YoutubeResult{}
		// yt-dlp args with custom output
		args := []string{
			"--print", "%(id)s|%(channel)s|%(title)s|%(duration_string)s|%(webpage_url)s|%(thumbnail)s",
			"--flat-playlist",
			"--skip-download",
			playlistURL,
		}
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

		lines := strings.Split(string(output), "\n")
		var i int64 = 1
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

func parseYoutubeOutput(output []byte) (YoutubeResult, error) {
	result := YoutubeResult{}

	line := strings.TrimSpace(string(output))
	if line == "" {
		return result, fmt.Errorf("empty output from yt-dlp")
	}

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