package player

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/mocks"
	"github.com/coreyo-git/beatgopher/services"
)

func TestPlayerAddSong(t *testing.T) {
	// Create mock dependencies
	mockSession := mocks.NewMockDiscordSession("test-guild", "test-channel")
	mockQueue := mocks.NewMockQueue()

	// Create player with mock dependencies
	player := &Player{
		Queue:         mockQueue,
		Session:       mockSession,
		CurrentStream: nil,
		IsPlaying:     false,
		stop:          make(chan bool),
	}

	// Create a test song
	testSong := &services.YoutubeResult{
		ID:        "test-song-id",
		Channel:   "Test Channel",
		Title:     "Test Song",
		Duration:  "3:30",
		URL:       "https://youtube.com/watch?v=test",
		Thumbnail: "test-thumbnail.jpg",
	}

	// Mock interaction (minimal setup for testing)
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID: "test-interaction",
		},
	}

	// Test adding a song
	player.AddSong(interaction, testSong)

	// Verify the song was added to the queue
	if mockQueue.Size() != 1 {
		t.Errorf("Expected queue size to be 1, got %d", mockQueue.Size())
	}

	// Verify the correct song was added
	queuedSong := mockQueue.Peek()
	if queuedSong == nil {
		t.Fatal("Expected a song in the queue, got nil")
	}

	if queuedSong.ID != testSong.ID {
		t.Errorf("Expected song ID %s, got %s", testSong.ID, queuedSong.ID)
	}

	if queuedSong.Title != testSong.Title {
		t.Errorf("Expected song title %s, got %s", testSong.Title, queuedSong.Title)
	}
}

func TestPlayerSkip(t *testing.T) {
	// Create mock dependencies
	mockSession := mocks.NewMockDiscordSession("test-guild", "test-channel")
	mockQueue := mocks.NewMockQueue()

	// Create player with mock dependencies
	player := &Player{
		Queue:         mockQueue,
		Session:       mockSession,
		CurrentStream: nil,
		IsPlaying:     false,
		stop:          make(chan bool, 1), // Buffered channel to avoid blocking
	}

	// Test skipping when not playing
	result := player.Skip()
	if result {
		t.Error("Expected Skip() to return false when not playing")
	}

	// Test skipping when playing
	player.IsPlaying = true

	// Skip should return true and send a signal
	result = player.Skip()
	if !result {
		t.Error("Expected Skip() to return true when playing")
	}

	// Verify that a signal was sent to the stop channel
	select {
	case <-player.stop:
		// Success - stop signal was received
	default:
		t.Error("Expected stop signal to be sent")
	}
}

func TestPlayerStop(t *testing.T) {
	// Create mock dependencies
	mockSession := mocks.NewMockDiscordSession("test-guild", "test-channel")
	mockQueue := mocks.NewMockQueue()

	// Add some test songs to the queue
	testSong1 := &services.YoutubeResult{ID: "song1", Title: "Song 1"}
	testSong2 := &services.YoutubeResult{ID: "song2", Title: "Song 2"}
	mockQueue.Enqueue(testSong1)
	mockQueue.Enqueue(testSong2)

	// Create player with mock dependencies
	player := &Player{
		Queue:         mockQueue,
		Session:       mockSession,
		CurrentStream: nil,
		IsPlaying:     true,
		stop:          make(chan bool),
	}

	// Verify queue has songs before stopping
	if mockQueue.Size() != 2 {
		t.Errorf("Expected queue size to be 2, got %d", mockQueue.Size())
	}

	// Stop the player
	player.Stop()

	// Verify player state after stopping
	if player.IsPlaying {
		t.Error("Expected IsPlaying to be false after Stop()")
	}

	// Verify queue was cleared (new queue is created)
	if !player.Queue.IsEmpty() {
		t.Error("Expected queue to be empty after Stop()")
	}
}

func TestPlayerGetters(t *testing.T) {
	// Create mock dependencies
	mockSession := mocks.NewMockDiscordSession("test-guild", "test-channel")
	mockQueue := mocks.NewMockQueue()

	// Create player with mock dependencies
	player := &Player{
		Queue:         mockQueue,
		Session:       mockSession,
		CurrentStream: nil,
		IsPlaying:     true,
		stop:          make(chan bool),
	}

	// Test GetQueue
	queue := player.GetQueue()
	if queue != mockQueue {
		t.Error("GetQueue() did not return the expected queue")
	}

	// Test GetSession
	session := player.GetSession()
	if session != mockSession {
		t.Error("GetSession() did not return the expected session")
	}

	// Test IsPlayerPlaying
	isPlaying := player.IsPlayerPlaying()
	if !isPlaying {
		t.Error("Expected IsPlayerPlaying() to return true")
	}

	// Change state and test again
	player.IsPlaying = false
	isPlaying = player.IsPlayerPlaying()
	if isPlaying {
		t.Error("Expected IsPlayerPlaying() to return false")
	}
}
