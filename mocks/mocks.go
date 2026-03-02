package mocks

import (
	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/services"
)

// MockDiscordSession is a mock implementation of DiscordSessionInterface for testing
type MockDiscordSession struct {
	guildID         string
	textChannelID   string
	voiceConnection *discordgo.VoiceConnection
	messages        []string
	embedsSent      []string
}

func NewMockDiscordSession(guildID, textChannelID string) *MockDiscordSession {
	return &MockDiscordSession{
		guildID:       guildID,
		textChannelID: textChannelID,
		messages:      []string{},
		embedsSent:    []string{},
	}
}

func (mds *MockDiscordSession) InteractionRespond(i *discordgo.Interaction, content string) error {
	mds.messages = append(mds.messages, content)
	return nil
}

func (mds *MockDiscordSession) FollowupMessage(i *discordgo.Interaction, content string) error {
	mds.messages = append(mds.messages, content)
	return nil
}

func (mds *MockDiscordSession) SendChannelMessage(message string) error {
	mds.messages = append(mds.messages, message)
	return nil
}

func (mds *MockDiscordSession) SendSongEmbed(song *services.YoutubeResult, footer string) error {
	mds.embedsSent = append(mds.embedsSent, song.Title+" - "+footer)
	return nil
}

func (mds *MockDiscordSession) SendQueueEmbed(songs []*services.YoutubeResult, currentPage int, totalPages int) error {
	mds.embedsSent = append(mds.embedsSent, "Queue embed sent")
	return nil
}

func (mds *MockDiscordSession) JoinVoiceChannel(i *discordgo.InteractionCreate) error {
	// Mock implementation - just return nil for success
	return nil
}

func (mds *MockDiscordSession) LeaveVoiceChannel() {
	// Mock implementation - do nothing
}

func (mds *MockDiscordSession) GetGuildID() string {
	return mds.guildID
}

func (mds *MockDiscordSession) GetTextChannelID() string {
	return mds.textChannelID
}

func (mds *MockDiscordSession) GetVoiceConnection() *discordgo.VoiceConnection {
	return mds.voiceConnection
}

// GetMessages returns all messages sent through this mock session
func (mds *MockDiscordSession) GetMessages() []string {
	return mds.messages
}

// GetEmbedsSent returns all embeds sent through this mock session
func (mds *MockDiscordSession) GetEmbedsSent() []string {
	return mds.embedsSent
}

// MockYoutubeService is a mock implementation of YoutubeServiceInterface for testing
type MockYoutubeService struct {
	searchResults map[string]services.YoutubeResult
	infoResults   map[string]services.YoutubeResult
}

func NewMockYoutubeService() *MockYoutubeService {
	return &MockYoutubeService{
		searchResults: make(map[string]services.YoutubeResult),
		infoResults:   make(map[string]services.YoutubeResult),
	}
}

func (mys *MockYoutubeService) GetYoutubeInfo(url string) (services.YoutubeResult, error) {
	if result, exists := mys.infoResults[url]; exists {
		return result, nil
	}
	// Return a default result for testing
	return services.YoutubeResult{
		ID:        "test-id",
		Channel:   "Test Channel",
		Title:     "Test Video",
		Duration:  "3:30",
		URL:       url,
		Thumbnail: "test-thumbnail.jpg",
	}, nil
}

func (mys *MockYoutubeService) SearchYoutube(query string) (services.YoutubeResult, error) {
	if result, exists := mys.searchResults[query]; exists {
		return result, nil
	}
	// Return a default result for testing
	return services.YoutubeResult{
		ID:        "search-test-id",
		Channel:   "Search Test Channel",
		Title:     "Search Test Video for: " + query,
		Duration:  "4:15",
		URL:       "https://youtube.com/watch?v=test",
		Thumbnail: "search-test-thumbnail.jpg",
	}, nil
}

func (mys *MockYoutubeService) GetYoutubePlaylistInfo(playlistURL string, total int64, randomizeSongs bool) ([]services.YoutubeResult, error) {
	// Return mock playlist results
	results := []services.YoutubeResult{
		{
			ID:        "playlist-song-1",
			Channel:   "Playlist Channel",
			Title:     "Playlist Song 1",
			Duration:  "3:45",
			URL:       "https://youtube.com/watch?v=playlist1",
			Thumbnail: "playlist1-thumbnail.jpg",
		},
		{
			ID:        "playlist-song-2",
			Channel:   "Playlist Channel",
			Title:     "Playlist Song 2",
			Duration:  "4:20",
			URL:       "https://youtube.com/watch?v=playlist2",
			Thumbnail: "playlist2-thumbnail.jpg",
		},
	}

	if total < int64(len(results)) {
		results = results[:total]
	}

	return results, nil
}

// SetSearchResult allows setting custom search results for testing
func (mys *MockYoutubeService) SetSearchResult(query string, result services.YoutubeResult) {
	mys.searchResults[query] = result
}

// SetInfoResult allows setting custom info results for testing
func (mys *MockYoutubeService) SetInfoResult(url string, result services.YoutubeResult) {
	mys.infoResults[url] = result
}

// Verify that MockYoutubeService implements YoutubeServiceInterface at compile time
var _ services.YoutubeServiceInterface = (*MockYoutubeService)(nil)
