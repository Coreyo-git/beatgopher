package player_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/coreyo-git/beatgopher/queue"
	"github.com/coreyo-git/beatgopher/player"
	"github.com/coreyo-git/beatgopher/services"
)

// createTestPlayer creates a Player with dependencies
func createTestPlayer(q queue.QueueInterface) *player.Player {

	return player.NewPlayer(q, func(song *services.YoutubeResult, content string) error {
		return nil
	},
		func() bool {
			return true
		},
		func() *discordgo.VoiceConnection {
			return nil
		},
		func() {},
	)
}

func TestPlayerAddSong(t *testing.T) {
	q := queue.NewQueue()
	player := createTestPlayer(q)
	player.IsPlaying = true

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
	if q.Size() != 1 {
		t.Errorf("Expected queue size to be 1, got %d", q.Size())
	}

	// Verify the correct song was added
	queuedSong := q.Peek()
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

func TestPlayerAddSongs(t *testing.T) {
	q := queue.NewQueue()
	player := createTestPlayer(q)
	player.IsPlaying = true
	playlist := []services.YoutubeResult{}
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID: "test-interaction",
		},
	}

	testSong1 := &services.YoutubeResult{
		ID:        "test-song-id1",
		Channel:   "Test Channel1",
		Title:     "Test Song1",
		Duration:  "3:31",
		URL:       "https://youtube.com/watch?v=test1",
		Thumbnail: "test-thumbnail.jpg1",
	}
	testSong2 := &services.YoutubeResult{
		ID:        "test-song-id2",
		Channel:   "Test Channel2",
		Title:     "Test Song2",
		Duration:  "3:32",
		URL:       "https://youtube.com/watch?v=test2",
		Thumbnail: "test-thumbnail.jpg2",
	}

	playlist = append(playlist, *testSong1)
	playlist = append(playlist, *testSong2)

	player.AddSongs(interaction, playlist)

	// Verify the song was added to the queue
	if q.Size() != 2 {
		t.Errorf("Expected queue size to be 2, got %d", q.Size())
	}

	// Verify the correct song was added
	queuedSong := q.Peek()
	if queuedSong == nil {
		t.Fatal("Expected test song 1 in the queue, got nil")
	}

	if queuedSong.ID != testSong1.ID {
		t.Errorf("Expected song ID %s, got %s", testSong1.ID, queuedSong.ID)
	}

	if queuedSong.Title != testSong1.Title {
		t.Errorf("Expected song title %s, got %s", testSong1.Title, queuedSong.Title)
	}

	// skip to test song 2
	q.Dequeue()

	// Verify the correct song was added
	queuedSong = q.Peek()
	if queuedSong == nil {
		t.Fatal("Expected test song 2 in the queue, got nil")
	}

	if queuedSong.ID != testSong2.ID {
		t.Errorf("Expected song ID %s, got %s", testSong2.ID, queuedSong.ID)
	}

	if queuedSong.Title != testSong2.Title {
		t.Errorf("Expected song title %s, got %s", testSong2.Title, queuedSong.Title)
	}
}

func TestPlayerSkip(t *testing.T) {
	q := queue.NewQueue()
	player := createTestPlayer(q)

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
}

func TestPlayerStop(t *testing.T) {
	q := queue.NewQueue()

	// Track if OnLeaveVoiceChannel was called
	leaveChannelCalled := false

	player := player.NewPlayer(q, func(song *services.YoutubeResult, content string) error {
		return nil
	},
		func() bool {
			return true
		},
		func() *discordgo.VoiceConnection {
			return nil
		},
		func() {
			leaveChannelCalled = true
		})

	// Add some test songs to the queue
	testSong1 := &services.YoutubeResult{ID: "song1", Title: "Song 1"}
	testSong2 := &services.YoutubeResult{ID: "song2", Title: "Song 2"}
	q.Enqueue(testSong1)
	q.Enqueue(testSong2)

	// Verify queue has songs before stopping
	if q.Size() != 2 {
		t.Errorf("Expected queue size to be 2, got %d", q.Size())
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
	q := queue.NewQueue()
	player := createTestPlayer(q)
	player.IsPlaying = true

	// Test GetQueue
	returnedQueue := player.GetQueue()
	if returnedQueue != q {
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
	q := queue.NewQueue()

	sendEmbedCalled := false
	checkVoiceCalled := false

	player := player.NewPlayer(
		q,
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

	if player.Queue != q {
		t.Error("Expected Queue to be the provided queue")
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
