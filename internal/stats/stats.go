// Package stats tracks per-request activity for the proxy so the TUI can show
// a live "recent activity" feed plus aggregate counters.
//
// The limiter keeps its own aggregate counters; this package supplements them
// with a bounded ring buffer of human-readable activity entries.
package stats

import (
	"sync"
	"time"
)

// Entry is a single activity record.
type Entry struct {
	Time     time.Time
	Method   string
	Host     string
	Path     string
	WaitedMs int64
	Status   int
	BytesIn  int64
	BytesOut int64
	Err      string
}

// Tracker keeps a bounded ring buffer of recent activity entries plus a few
// convenience aggregates that are useful to the TUI.
type Tracker struct {
	mu sync.Mutex

	buf  []Entry
	head int // index of the oldest entry
	size int // number of live entries
	cap  int

	totalBytesIn  int64
	totalBytesOut int64
	errCount      int64
}

// New constructs a Tracker holding at most `cap` recent entries.
func New(cap int) *Tracker {
	if cap < 8 {
		cap = 8
	}
	return &Tracker{
		buf: make([]Entry, cap),
		cap: cap,
	}
}

// Push appends an entry, evicting the oldest when full.
func (t *Tracker) Push(e Entry) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.buf[t.head] = e
	t.head = (t.head + 1) % t.cap
	if t.size < t.cap {
		t.size++
	}
	t.totalBytesIn += e.BytesIn
	t.totalBytesOut += e.BytesOut
	if e.Err != "" || e.Status >= 400 {
		t.errCount++
	}
}

// Recent returns up to n entries, newest first.
func (t *Tracker) Recent(n int) []Entry {
	t.mu.Lock()
	defer t.mu.Unlock()

	if n <= 0 || n > t.size {
		n = t.size
	}
	out := make([]Entry, 0, n)
	// Newest is at index (head-1) mod cap, going backwards.
	for i := 0; i < n; i++ {
		idx := (t.head - 1 - i + t.cap) % t.cap
		out = append(out, t.buf[idx])
	}
	return out
}

// Totals returns aggregate counters.
func (t *Tracker) Totals() (bytesIn, bytesOut, errs int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.totalBytesIn, t.totalBytesOut, t.errCount
}

// Reset clears the ring buffer and the aggregates.
func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.head = 0
	t.size = 0
	t.totalBytesIn = 0
	t.totalBytesOut = 0
	t.errCount = 0
}
