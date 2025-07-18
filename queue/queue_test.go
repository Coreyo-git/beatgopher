package queue

import (
	"testing"

	"github.com/coreyo-git/beatgopher/services"
)

func TestQueueOperations(t *testing.T) {
	q := &Queue{
		songs: []*services.YoutubeResult{},
	}

	// Test IsEmpty on an empty queue
	if !q.IsEmpty() {
		t.Error("Expected queue to be empty, but it was not")
	}

	// Test Enqueue
	song1 := services.YoutubeResult{
		ID: "1", 
		Channel: "Channel 1",
		Title: "Song 1",
		Duration: "1:00", 
		URL: "Song1.com",
		Thumbnail: "Song1Thumbnail.link",

	}
	song2 := services.YoutubeResult{
		ID: "2", 
		Channel: "Channel 2",
		Title: "Song 2",
		Duration: "2:00", 
		URL: "Song2.com",
		Thumbnail: "Song2Thumbnail.link",

	}
	song3 := services.YoutubeResult{
		ID: "3", 
		Channel: "Channel 3",
		Title: "Song 3",
		Duration: "3:00", 
		URL: "Song3.com",
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