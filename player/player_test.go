package player

import (
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/mocks"
	"github.com/coreyo-git/beatgopher/services"
)

// createTestPlayer creates a Player with mock dependencies and no-op callbacks
func createTestPlayer(mockQueue *mocks.MockQueue) *Player {
	return &Player{
		Queue:         mockQueue,
		CurrentStream: nil,
		IsPlaying:     false,
		stop:          make(chan bool, 1),
		skip:          make(chan bool, 1),
		mu:            sync.RWMutex{},
		OnSendEmbedMessage: func(song *services.YoutubeResult, content string) error {
			return nil
		},
		OnCheckVoiceConnection: func() bool {
			return true
		},
		OnGetVoiceConnection: func() *discordgo.VoiceConnection {
			return nil
		},
		OnLeaveVoiceChannel: func() {},
	}
}

func TestPlayerAddSong(t *testing.T) {
	mockQueue := mocks.NewMockQueue()
	player := createTestPlayer(mockQueue)

	testSong := &services.YoutubeResult{
		ID:        "test-song-id",
		Channel:   "Test Channel",
		Title:     "Test Song",
		Duration:  "3:30",
		URL:       "https://youtube.com/watch?v=test",
		Thumbnail: "test-thumbnail.jpg",
	}

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID: "test-interaction",
		},
	}

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
	mockQueue := mocks.NewMockQueue()
	player := createTestPlayer(mockQueue)

	// Test skipping when not playing
	result := player.Skip()
	if result {
		t.Error("Expected Skip() to return false when not playing")
	}

	// Test skipping when playing
	player.IsPlaying = true

	result = player.Skip()
	if !result {
		t.Error("Expected Skip() to return true when playing")
	}

	// Verify that a signal was sent to the skip channel (not stop)
	select {
	case <-player.skip:
		// Success - skip signal was received
	default:
		t.Error("Expected skip signal to be sent")
	}
}

func TestPlayerStop(t *testing.T) {
	mockQueue := mocks.NewMockQueue()

	// Track if OnLeaveVoiceChannel was called
	leaveChannelCalled := false

	player := &Player{
		Queue:         mockQueue,
		CurrentStream: nil,
		IsPlaying:     true,
		stop:          make(chan bool, 1),
		skip:          make(chan bool, 1),
		mu:            sync.RWMutex{},
		OnSendEmbedMessage: func(song *services.YoutubeResult, content string) error {
			return nil
		},
		OnCheckVoiceConnection: func() bool {
			return true
		},
		OnGetVoiceConnection: func() *discordgo.VoiceConnection {
			return nil
		},
		OnLeaveVoiceChannel: func() {
			leaveChannelCalled = true
		},
	}

	// Add some test songs to the queue
	testSong1 := &services.YoutubeResult{ID: "song1", Title: "Song 1"}
	testSong2 := &services.YoutubeResult{ID: "song2", Title: "Song 2"}
	mockQueue.Enqueue(testSong1)
	mockQueue.Enqueue(testSong2)

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

	// Verify OnLeaveVoiceChannel was called
	if !leaveChannelCalled {
		t.Error("Expected OnLeaveVoiceChannel to be called")
	}
}

func TestPlayerIsPlayerPlaying(t *testing.T) {
	mockQueue := mocks.NewMockQueue()
	player := createTestPlayer(mockQueue)
	player.IsPlaying = true

	// Test GetQueue
	queue := player.GetQueue()
	if queue != mockQueue {
		t.Error("GetQueue() did not return the expected queue")
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

func TestNewPlayer(t *testing.T) {
	mockQueue := mocks.NewMockQueue()

	sendEmbedCalled := false
	checkVoiceCalled := false

	player := NewPlayer(
		mockQueue,
		func(song *services.YoutubeResult, content string) error {
			sendEmbedCalled = true
			return nil
		},
		func() bool {
			checkVoiceCalled = true
			return true
		},
		func() *discordgo.VoiceConnection {
			return nil
		},
		func() {},
	)

	// Verify initial state
	if player.IsPlaying {
		t.Error("Expected IsPlaying to be false initially")
	}

	if player.CurrentStream != nil {
		t.Error("Expected CurrentStream to be nil initially")
	}

	if player.Queue != mockQueue {
		t.Error("Expected Queue to be the provided mock queue")
	}

	// Verify callbacks are wired correctly
	player.OnSendEmbedMessage(&services.YoutubeResult{}, "test")
	if !sendEmbedCalled {
		t.Error("Expected OnSendEmbedMessage callback to be called")
	}

	player.OnCheckVoiceConnection()
	if !checkVoiceCalled {
		t.Error("Expected OnCheckVoiceConnection callback to be called")
	}
}
