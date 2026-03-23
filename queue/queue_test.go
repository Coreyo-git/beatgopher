package queue

import (
	"fmt"
	"sync"
	"testing"

	"github.com/coreyo-git/beatgopher/services"
)

func TestQueueOperations(t *testing.T) {
	q := NewQueue()

	// Test IsEmpty on an empty queue
	if !q.IsEmpty() {
		t.Error("Expected queue to be empty, but it was not")
	}

	// Test Enqueue
	song1 := services.YoutubeResult{
		ID:        "1",
		Channel:   "Channel 1",
		Title:     "Song 1",
		Duration:  "1:00",
		URL:       "Song1.com",
		Thumbnail: "Song1Thumbnail.link",
	}
	song2 := services.YoutubeResult{
		ID:        "2",
		Channel:   "Channel 2",
		Title:     "Song 2",
		Duration:  "2:00",
		URL:       "Song2.com",
		Thumbnail: "Song2Thumbnail.link",
	}
	song3 := services.YoutubeResult{
		ID:        "3",
		Channel:   "Channel 3",
		Title:     "Song 3",
		Duration:  "3:00",
		URL:       "Song3.com",
		Thumbnail: "Song3Thumbnail.link",
	}

	q.Enqueue(&song1)
	if q.IsEmpty() {
		t.Error("Expected queue not to be empty after enqueue")
	}

	q.Enqueue(&song2)
	q.Enqueue(&song3)
	if q.Size() != 3 {
		t.Errorf("Expected size 3, got %d", q.Size())
	}

	// Test Peek
	peekedItem := *q.Peek()
	if peekedItem != song1 {
		t.Errorf("Expected peeked item to be Song 1, got %v", peekedItem)
	}
	if q.Size() != 3 { // Peek should not change size
		t.Errorf("Expected size 3 after peek, got %d", q.Size())
	}

	// Test Dequeue
	dequeuedItem := *q.Dequeue()
	if dequeuedItem != song1 {
		t.Errorf("Expected dequeued item to be Song 1, got %v", dequeuedItem)
	}
	if q.Size() != 2 {
		t.Errorf("Expected size 2 after dequeue, got %d", q.Size())
	}

	dequeuedItem = *q.Dequeue()
	if dequeuedItem != song2 {
		t.Errorf("Expected dequeued item to be Song 2, got %v", dequeuedItem)
	}
	if q.Size() != 1 {
		t.Errorf("Expected size 1 after dequeue, got %d", q.Size())
	}

	dequeuedItem = *q.Dequeue()
	if dequeuedItem != song3 {
		t.Errorf("Expected dequeued item to be Song 3, got %v", dequeuedItem)
	}
	if q.Size() != 0 {
		t.Errorf("Expected size 0 after dequeue, got %d", q.Size())
	}
	if !q.IsEmpty() {
		t.Error("Expected queue to be empty after all items dequeued")
	}

	// Test Dequeue on empty queue
	dequeuedItemPtr := q.Dequeue()
	if dequeuedItemPtr != nil {
		t.Errorf("Expected nil when dequeuing from empty queue, got %v", dequeuedItem)
	}
}

func TestRemoveFromQueue(t *testing.T) {
	q := NewQueue()

	// Create test songs
	song1 := &services.YoutubeResult{
		ID:        "test1",
		Channel:   "Test Channel 1",
		Title:     "Test Song 1",
		Duration:  "3:30",
		URL:       "https://youtube.com/watch?v=test1",
		Thumbnail: "https://img.youtube.com/vi/test1/default.jpg",
	}

	song2 := &services.YoutubeResult{
		ID:        "test2",
		Channel:   "Test Channel 2",
		Title:     "Test Song 2",
		Duration:  "4:15",
		URL:       "https://youtube.com/watch?v=test2",
		Thumbnail: "https://img.youtube.com/vi/test2/default.jpg",
	}

	song3 := &services.YoutubeResult{
		ID:        "test3",
		Channel:   "Test Channel 3",
		Title:     "Test Song 3",
		Duration:  "2:45",
		URL:       "https://youtube.com/watch?v=test3",
		Thumbnail: "https://img.youtube.com/vi/test3/default.jpg",
	}

	// Test removing from empty queue
	if q.RemoveFromQueue(song1) {
		t.Error("Expected RemoveFromQueue to return false for empty queue")
	}

	// Add songs to queue
	q.Enqueue(song1)
	q.Enqueue(song2)
	q.Enqueue(song3)

	if q.Size() != 3 {
		t.Errorf("Expected queue size 3, got %d", q.Size())
	}

	// Test removing middle song
	if !q.RemoveFromQueue(song2) {
		t.Error("Expected RemoveFromQueue to return true for existing song")
	}

	if q.Size() != 2 {
		t.Errorf("Expected queue size 2 after removal, got %d", q.Size())
	}

	// Verify the correct songs remain and in correct order
	songs := q.GetSongs()
	if len(songs) != 2 {
		t.Errorf("Expected 2 songs in GetSongs(), got %d", len(songs))
	}

	if songs[0].ID != song1.ID {
		t.Errorf("Expected first song to be %s, got %s", song1.ID, songs[0].ID)
	}

	if songs[1].ID != song3.ID {
		t.Errorf("Expected second song to be %s, got %s", song3.ID, songs[1].ID)
	}

	// Test removing first song
	if !q.RemoveFromQueue(song1) {
		t.Error("Expected RemoveFromQueue to return true for first song")
	}

	if q.Size() != 1 {
		t.Errorf("Expected queue size 1 after removal, got %d", q.Size())
	}

	// Verify only song3 remains
	if q.Peek().ID != song3.ID {
		t.Errorf("Expected remaining song to be %s, got %s", song3.ID, q.Peek().ID)
	}

	// Test removing last song
	if !q.RemoveFromQueue(song3) {
		t.Error("Expected RemoveFromQueue to return true for last song")
	}

	if !q.IsEmpty() {
		t.Error("Expected queue to be empty after removing all songs")
	}

	// Test removing non-existent song
	nonExistentSong := &services.YoutubeResult{
		ID:        "nonexistent",
		Channel:   "Non-existent Channel",
		Title:     "Non-existent Song",
		Duration:  "0:00",
		URL:       "https://youtube.com/watch?v=nonexistent",
		Thumbnail: "https://img.youtube.com/vi/nonexistent/default.jpg",
	}

	if q.RemoveFromQueue(nonExistentSong) {
		t.Error("Expected RemoveFromQueue to return false for non-existent song")
	}
}

func TestQueueWithFilledQueue(t *testing.T) {
	q := NewQueue()

	// Create a filled queue with multiple songs
	songs := []*services.YoutubeResult{
		{
			ID:        "fill1",
			Channel:   "Channel A",
			Title:     "First Song",
			Duration:  "3:00",
			URL:       "https://youtube.com/watch?v=fill1",
			Thumbnail: "https://img.youtube.com/vi/fill1/default.jpg",
		},
		{
			ID:        "fill2",
			Channel:   "Channel B",
			Title:     "Second Song",
			Duration:  "4:00",
			URL:       "https://youtube.com/watch?v=fill2",
			Thumbnail: "https://img.youtube.com/vi/fill2/default.jpg",
		},
		{
			ID:        "fill3",
			Channel:   "Channel C",
			Title:     "Third Song",
			Duration:  "2:30",
			URL:       "https://youtube.com/watch?v=fill3",
			Thumbnail: "https://img.youtube.com/vi/fill3/default.jpg",
		},
		{
			ID:        "fill4",
			Channel:   "Channel D",
			Title:     "Fourth Song",
			Duration:  "5:15",
			URL:       "https://youtube.com/watch?v=fill4",
			Thumbnail: "https://img.youtube.com/vi/fill4/default.jpg",
		},
		{
			ID:        "fill5",
			Channel:   "Channel E",
			Title:     "Fifth Song",
			Duration:  "3:45",
			URL:       "https://youtube.com/watch?v=fill5",
			Thumbnail: "https://img.youtube.com/vi/fill5/default.jpg",
		},
	}

	// Add all songs to queue
	for _, song := range songs {
		q.Enqueue(song)
	}

	// Test initial state
	if q.Size() != 5 {
		t.Errorf("Expected queue size 5, got %d", q.Size())
	}

	if q.IsEmpty() {
		t.Error("Expected queue not to be empty")
	}

	// Test removing from different positions
	// Remove middle song (index 2, "Third Song")
	if !q.RemoveFromQueue(songs[2]) {
		t.Error("Failed to remove middle song")
	}

	if q.Size() != 4 {
		t.Errorf("Expected queue size 4 after removing middle song, got %d", q.Size())
	}

	// Verify order is maintained
	remainingSongs := q.GetSongs()
	expectedIDs := []string{"fill1", "fill2", "fill4", "fill5"}

	for i, expectedID := range expectedIDs {
		if remainingSongs[i].ID != expectedID {
			t.Errorf("Expected song at position %d to have ID %s, got %s", i, expectedID, remainingSongs[i].ID)
		}
	}

	// Remove first song
	if !q.RemoveFromQueue(songs[0]) {
		t.Error("Failed to remove first song")
	}

	if q.Size() != 3 {
		t.Errorf("Expected queue size 3 after removing first song, got %d", q.Size())
	}

	// Verify first song is now "Second Song"
	if q.Peek().ID != "fill2" {
		t.Errorf("Expected first song to be fill2, got %s", q.Peek().ID)
	}

	// Remove last song
	if !q.RemoveFromQueue(songs[4]) {
		t.Error("Failed to remove last song")
	}

	if q.Size() != 2 {
		t.Errorf("Expected queue size 2 after removing last song, got %d", q.Size())
	}

	// Verify remaining songs
	finalSongs := q.GetSongs()
	expectedFinalIDs := []string{"fill2", "fill4"}

	for i, expectedID := range expectedFinalIDs {
		if finalSongs[i].ID != expectedID {
			t.Errorf("Expected final song at position %d to have ID %s, got %s", i, expectedID, finalSongs[i].ID)
		}
	}
}

func TestConcurrentQueueOperations(t *testing.T) {
	q := NewQueue()
	var wg sync.WaitGroup

	wg.Add(2)
	// Goroutine 1: Add songs
	wg.Go(func() {
		for i := 0; i < 10; i++ {
			testSong := &services.YoutubeResult{
				ID:        fmt.Sprintf("conc%d", i),
				Channel:   "Test Channel",
				Title:     fmt.Sprintf("Test Song %d", i),
				Duration:  "3:00",
				URL:       fmt.Sprintf("https://youtube.com/watch?v=conc%d", i),
				Thumbnail: fmt.Sprintf("https://img.youtube.com/vi/conc%d/default.jpg", i),
			}
			q.Enqueue(testSong)
		}
		wg.Done()
	})

	wg.Go(func() {
		for i := 10; i < 20; i++ {
			testSong := &services.YoutubeResult{
				ID:        fmt.Sprintf("conc%d", i),
				Channel:   "Test Channel",
				Title:     fmt.Sprintf("Test Song %d", i),
				Duration:  "3:00",
				URL:       fmt.Sprintf("https://youtube.com/watch?v=conc%d", i),
				Thumbnail: fmt.Sprintf("https://img.youtube.com/vi/conc%d/default.jpg", i),
			}
			q.Enqueue(testSong)
		}
		wg.Done()
	})

	wg.Wait()

	// Test Size and no duplicate ID's
	size := q.Size()
	if size != 20 {
		t.Errorf("Expected 20 songs, actually got %d", size)
	}

	songs := q.GetSongs()
	seen := make(map[string]bool)
	for _, song := range songs {
		if seen[song.ID] {
			t.Errorf("Duplicate song ID Found: %s", song.ID)
		}
		seen[song.ID] = true
	}

	q.Clear()

	// Add songs to test dequeue concurrency
	numSongs := 30

	for i := 0; i < numSongs; i++ {
		testSong := &services.YoutubeResult{
			ID:        fmt.Sprintf("conc%d", i),
			Channel:   "Test Channel",
			Title:     fmt.Sprintf("Test Song %d", i),
			Duration:  "3:00",
			URL:       fmt.Sprintf("https://youtube.com/watch?v=conc%d", i),
			Thumbnail: fmt.Sprintf("https://img.youtube.com/vi/conc%d/default.jpg", i),
		}
		q.Enqueue(testSong)
	}

	songsRemoved := make([]services.YoutubeResult, numSongs)
	wg.Add(3)
	// Goroutine 2: Remove songs
	wg.Go(func() {
		for i := 0; i < 10; i++ {
			song := q.Dequeue()
			songsRemoved[i] = *song
		}
		wg.Done()
	})
	wg.Go(func() {
		for i := 10; i < 20; i++ {
			song := q.Dequeue()
			songsRemoved[i] = *song
		}
		wg.Done()
	})
	wg.Go(func() {
		for i := 20; i < 30; i++ {
			song := q.Dequeue()
			songsRemoved[i] = *song
		}
		wg.Done()
	})

	wg.Wait()
	// Verify slice is in a valid state
	size = q.Size()
	if size != 0 {
		t.Errorf("Expected queue size to be 0 but was actually %d", size)
	}

	lengthOfSongsRemoved := len(songsRemoved)
	if lengthOfSongsRemoved != 30 {
		t.Errorf("Expected number of songs removed was %d, actual: %d", numSongs, lengthOfSongsRemoved)
	}

	seen = make(map[string]bool)
	for _, song := range songsRemoved {
		if seen[song.ID] {
			t.Errorf("Duplicate song ID Found in songsRemoved slice: %s", song.ID)
		}
		seen[song.ID] = true
	}
}

func TestQueueIsEmptyOnClear(t *testing.T) {
	q := NewQueue()
	song1 := services.YoutubeResult{
		ID:        "1",
		Channel:   "Channel 1",
		Title:     "Song 1",
		Duration:  "1:00",
		URL:       "Song1.com",
		Thumbnail: "Song1Thumbnail.link",
	}
	song2 := services.YoutubeResult{
		ID:        "2",
		Channel:   "Channel 2",
		Title:     "Song 2",
		Duration:  "2:00",
		URL:       "Song2.com",
		Thumbnail: "Song2Thumbnail.link",
	}
	song3 := services.YoutubeResult{
		ID:        "3",
		Channel:   "Channel 3",
		Title:     "Song 3",
		Duration:  "3:00",
		URL:       "Song3.com",
		Thumbnail: "Song3Thumbnail.link",
	}

	q.Enqueue(&song1)
	q.Enqueue(&song2)
	q.Enqueue(&song3)

	q.Clear()

	if !q.IsEmpty() {
		t.Errorf("Expected queue to be empty after being clear.")
	}
}

func TestPeekOnEmptyQueue(t *testing.T) {
	q := NewQueue()

	result := q.Peek()

	if result != nil {
		t.Errorf("Expected queue peek to return nil on empty queue")
	}
}

// hasDuplicates checks if a slice of comparable elements has any duplicates.
func hasDuplicates[T comparable](slice []T) bool {
	// A map is used as a set, with the element type as the key and
	// an empty struct{} as the value for memory efficiency.
	seen := make(map[T]struct{})

	for _, value := range slice {
		if _, exists := seen[value]; exists {
			// A duplicate was found.
			return true
		}
		// Mark the element as seen.
		seen[value] = struct{}{}
	}

	// No duplicates were found after checking all elements.
	return false
}
