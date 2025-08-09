package queue

import "github.com/coreyo-git/beatgopher/services"

// QueueInterface defines the contract for queue operations
type QueueInterface interface {
	// Enqueue adds a song to the queue
	Enqueue(song *services.YoutubeResult)

	// Dequeue removes and returns the first song from the queue
	Dequeue() *services.YoutubeResult

	// RemoveFromQueue removes a specific song from the queue
	RemoveFromQueue(song *services.YoutubeResult) bool

	// IsEmpty returns true if the queue is empty
	IsEmpty() bool

	// Peek returns the first song without removing it
	Peek() *services.YoutubeResult

	// Size returns the number of songs in the queue
	Size() int

	// GetSongs returns a copy of all songs in the queue
	GetSongs() []*services.YoutubeResult
}

// Verify that Queue implements QueueInterface at compile time
var _ QueueInterface = (*Queue)(nil)
