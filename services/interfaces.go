package services

// YoutubeServiceInterface defines the contract for YouTube operations
type YoutubeServiceInterface interface {
	// GetYoutubeInfo gets information about a YouTube video from its URL
	GetYoutubeInfo(url string) (YoutubeResult, error)

	// SearchYoutube searches for a YouTube video by query
	SearchYoutube(query string) (YoutubeResult, error)

	// GetYoutubePlaylistInfo gets information about a YouTube playlist
	GetYoutubePlaylistInfo(playlistURL string, total int64, randomizeSongs bool) ([]YoutubeResult, error)
}

// AudioStreamInterface defines the contract for audio streaming operations
type AudioStreamInterface interface {
	// NewAudioStream creates a new audio stream from a URL
	NewAudioStream(url string) (*AudioStream, error)
	// Close closes the audio stream
	Close()
}

// AudioStreamProvider is a concrete implementation of AudioStreamInterface
type AudioStreamProvider struct{}

// NewAudioStream creates a new audio stream from a URL
func (asp *AudioStreamProvider) NewAudioStream(url string) (*AudioStream, error) {
	return NewAudioStream(url)
}

// Close closes the audio stream
func (asp *AudioStreamProvider) Close() {
	// This is a dummy implementation to satisfy the interface
	// The actual implementation is in the AudioStream struct
}

// YoutubeService is a concrete implementation of YoutubeServiceInterface
type YoutubeService struct{}

// GetYoutubeInfo gets information about a YouTube video from its URL
func (ys *YoutubeService) GetYoutubeInfo(url string) (YoutubeResult, error) {
	return GetYoutubeInfo(url)
}

// SearchYoutube searches for a YouTube video by query
func (ys *YoutubeService) SearchYoutube(query string) (YoutubeResult, error) {
	return SearchYoutube(query)
}

// GetYoutubePlaylistInfo gets information about a YouTube playlist
func (ys *YoutubeService) GetYoutubePlaylistInfo(playlistURL string, total int64, randomizeSongs bool) ([]YoutubeResult, error) {
	return GetYoutubePlaylistInfo(playlistURL, total, randomizeSongs)
}

// Verify that YoutubeService implements YoutubeServiceInterface at compile time
var _ YoutubeServiceInterface = (*YoutubeService)(nil)

// Verify that AudioStreamProvider implements AudioStreamInterface at compile time
var _ AudioStreamInterface = (*AudioStreamProvider)(nil)
