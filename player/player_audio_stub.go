//go:build !cgo

package player

import(
	"fmt"
	"io"

	"github.com/coreyo-git/beatgopher/services"
)

func stream(p *Player) {}

func setupAudioOutput(result *services.YoutubeResult, p *Player) (io.ReadCloser, error) {
	return nil, fmt.Errorf("audio unavailable: CGO required")
}
