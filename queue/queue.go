package queue

import (
	"sync"
	"log"
	"github.com/coreyo-git/beatgopher/services"
)

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

	// Clear removes all songs from the queue
	Clear()
}

// FIFO queue for a single guild
type Queue struct {
	mu sync.Mutex // Protects queue from 
	songs []*services.YoutubeResult
}

func NewQueue() *Queue {
	return &Queue{
		mu:    sync.Mutex{},
		songs: []*services.YoutubeResult{},
	}
}

func (q *Queue) Enqueue(song *services.YoutubeResult) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.songs = append(q.songs, song)
}

func (q *Queue) Dequeue () *services.YoutubeResult {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.songs) == 0 {
		log.Printf("Queue is empty.")
		return nil
	}

	song := q.songs[0]
	q.songs = q.songs[1:]
	return song
}

func (q *Queue) RemoveFromQueue(song *services.YoutubeResult) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.songs) == 0 {
		log.Printf("Cannot Remove From an empty Queue.")
		return false
	}
	for i := range q.songs {
		if (q.songs)[i].ID == song.ID {
			// Remove the element by slicing and appending
			// This creates a new slice without the element at i
			q.songs = append((q.songs)[:i], (q.songs)[i+1:]...)
			return true
		}
	}
	log.Printf("Could not find song in queue.")
	return false
}

func (q *Queue) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.songs) == 0
}

func (q *Queue) Peek() *services.YoutubeResult {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.songs) == 0 {
		return nil
	}
	return q.songs[0]
}

func (q *Queue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.songs)
}

func (q *Queue) GetSongs() []*services.YoutubeResult {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Return a copy of the slice to avoid race conditions
	songsCopy := make([]*services.YoutubeResult, len(q.songs))
	copy(songsCopy, q.songs)
	return songsCopy
}

func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.songs = []*services.YoutubeResult{}
}
