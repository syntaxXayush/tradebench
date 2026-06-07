package runner

import (
	"sync"
	"time"
)

// Entry holds runtime info for a live contestant container.
type Entry struct {
	ContainerID string
	Port        int
	StartedAt   time.Time
}

// Registry is a concurrency-safe map of submissionID → Entry.
// It is written by the consumer goroutine and read by the watchdog goroutine.
type Registry struct {
	mu      sync.Mutex
	entries map[string]Entry
}

func NewRegistry() *Registry {
	return &Registry{entries: make(map[string]Entry)}
}

func (r *Registry) Add(submissionID, containerID string, port int, startedAt time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[submissionID] = Entry{ContainerID: containerID, Port: port, StartedAt: startedAt}
}

func (r *Registry) Remove(submissionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, submissionID)
}

// Snapshot returns a copy of all current entries for safe iteration.
func (r *Registry) Snapshot() map[string]Entry {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]Entry, len(r.entries))
	for k, v := range r.entries {
		out[k] = v
	}
	return out
}
