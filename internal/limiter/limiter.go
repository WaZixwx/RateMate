// Package limiter implements the rate-limiting / delay-injection core.
//
// Two operating modes:
//
//   - Auto:   the user supplies a time window (second/minute/hour) and a
//             maximum number of requests allowed within that window. The
//             limiter computes the inter-request spacing delay as
//             window / limit and enforces even spacing between successive
//             forwarded requests.
//
//   - Manual: the user supplies a fixed delay (ms) applied before every
//             request is forwarded.
//
// The limiter is safe for concurrent use.
package limiter

import (
	"context"
	"sync"
	"time"
)

// Mode selects the strategy used to compute the per-request delay.
type Mode int

const (
	// ModeAuto computes the delay from window + limit.
	ModeAuto Mode = iota
	// ModeManual uses a user-supplied millisecond delay.
	ModeManual
)

// String returns a human readable mode name.
func (m Mode) String() string {
	switch m {
	case ModeAuto:
		return "Auto"
	case ModeManual:
		return "Manual"
	default:
		return "Unknown"
	}
}

// Window is the time window used by ModeAuto.
type Window int

const (
	WindowSecond Window = iota
	WindowMinute
	WindowHour
)

// String returns a human readable window name.
func (w Window) String() string {
	switch w {
	case WindowSecond:
		return "Second"
	case WindowMinute:
		return "Minute"
	case WindowHour:
		return "Hour"
	default:
		return "Unknown"
	}
}

// Short returns a short label suitable for the TUI (s/m/h).
func (w Window) Short() string {
	switch w {
	case WindowSecond:
		return "s"
	case WindowMinute:
		return "m"
	case WindowHour:
		return "h"
	default:
		return "?"
	}
}

// Milliseconds returns the window length in milliseconds.
func (w Window) Milliseconds() int64 {
	switch w {
	case WindowSecond:
		return 1000
	case WindowMinute:
		return 60 * 1000
	case WindowHour:
		return 60 * 60 * 1000
	default:
		return 1000
	}
}

// Config holds the parameters consumed by the limiter.
//
// The limiter keeps its own copy of these fields; the TUI mutates them via
// Set* methods which are concurrency-safe.
type Config struct {
	Mode          Mode
	Window        Window
	Limit         int    // requests per window (Auto)
	ManualDelayMs int    // delay per request (Manual)
	Burst         int    // allowed burst (default 1) — kept >0 to keep semantics simple
}

// Limiter enforces the configured delay between forwarded requests.
type Limiter struct {
	mu sync.Mutex

	cfg Config

	// lastForward is the time the most recent request was released.
	lastForward time.Time

	// spacingMs is the cached delay currently in effect (computed under the
	// lock whenever the config changes).
	spacingMs int64

	// live counters, readable via Snapshot
	totalRequests int64
	delayedCount   int64
	passThrough    int64
	totalWaitedMs  int64
	lastRequestAt  time.Time

	// paused disables delaying entirely (requests pass straight through).
	paused bool
}

// New constructs a Limiter with the supplied initial config.
func New(cfg Config) *Limiter {
	l := &Limiter{cfg: cfg}
	l.recomputeSpacing()
	return l
}

// Config returns a snapshot of the current config.
func (l *Limiter) Config() Config {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cfg
}

// SetMode updates the operating mode and recomputes the spacing.
func (l *Limiter) SetMode(m Mode) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cfg.Mode = m
	l.recomputeSpacing()
}

// SetWindow updates the window and recomputes the spacing.
func (l *Limiter) SetWindow(w Window) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cfg.Window = w
	l.recomputeSpacing()
}

// SetLimit updates the request limit and recomputes the spacing.
func (l *Limiter) SetLimit(n int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if n < 1 {
		n = 1
	}
	l.cfg.Limit = n
	l.recomputeSpacing()
}

// SetManualDelay updates the manual delay (ms).
func (l *Limiter) SetManualDelay(ms int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if ms < 0 {
		ms = 0
	}
	l.cfg.ManualDelayMs = ms
	l.recomputeSpacing()
}

// SetPaused toggles pass-through mode. When paused, Wait returns immediately
// and requests are not delayed (but are still counted).
func (l *Limiter) SetPaused(p bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.paused = p
}

// Paused reports whether the limiter is in pass-through mode.
func (l *Limiter) Paused() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.paused
}

// recomputeSpacing must be called with the lock held.
func (l *Limiter) recomputeSpacing() {
	switch l.cfg.Mode {
	case ModeAuto:
		if l.cfg.Limit > 0 {
			l.spacingMs = l.cfg.Window.Milliseconds() / int64(l.cfg.Limit)
		} else {
			l.spacingMs = 0
		}
	case ModeManual:
		l.spacingMs = int64(l.cfg.ManualDelayMs)
	}
}

// SpacingMs returns the currently effective inter-request delay in ms.
func (l *Limiter) SpacingMs() int64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.spacingMs
}

// Wait blocks until the limiter allows the next request to proceed.
//
// It returns the duration it actually slept so callers can record statistics.
// If the context is cancelled while waiting, Wait returns ctx.Err() and a zero
// sleep duration.
func (l *Limiter) Wait(ctx context.Context) (slept time.Duration, err error) {
	l.mu.Lock()
	spacing := time.Duration(l.spacingMs) * time.Millisecond
	paused := l.paused
	last := l.lastForward

	// Record this request immediately so concurrent callers queue behind it.
	l.totalRequests++
	l.mu.Unlock()

	if paused || spacing <= 0 {
		l.mu.Lock()
		l.passThrough++
		l.lastForward = time.Now()
		l.lastRequestAt = l.lastForward
		l.mu.Unlock()
		return 0, nil
	}

	// Compute required wait relative to the previous release point.
	now := time.Now()
	var wait time.Duration
	if last.IsZero() {
		wait = 0
	} else {
		elapsed := now.Sub(last)
		if elapsed < spacing {
			wait = spacing - elapsed
		}
	}

	// Reserve the release slot immediately so the next caller queues behind
	// this one. This is what guarantees even spacing under concurrency.
	l.mu.Lock()
	l.lastForward = now.Add(wait)
	l.lastRequestAt = l.lastForward
	l.mu.Unlock()

	if wait > 0 {
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return 0, ctx.Err()
		}
		l.mu.Lock()
		l.delayedCount++
		l.totalWaitedMs += wait.Milliseconds()
		l.mu.Unlock()
	}
	return wait, nil
}

// Snapshot captures a consistent view of the live counters.
type Snapshot struct {
	TotalRequests int64
	DelayedCount  int64
	PassThrough   int64
	TotalWaitedMs int64
	LastRequestAt time.Time
	SpacingMs     int64
	Mode          Mode
	Window        Window
	Limit         int
	ManualDelayMs int
	Paused        bool
}

// Snapshot returns a consistent view of the limiter's state and counters.
func (l *Limiter) Snapshot() Snapshot {
	l.mu.Lock()
	defer l.mu.Unlock()
	return Snapshot{
		TotalRequests: l.totalRequests,
		DelayedCount:  l.delayedCount,
		PassThrough:   l.passThrough,
		TotalWaitedMs: l.totalWaitedMs,
		LastRequestAt: l.lastRequestAt,
		SpacingMs:     l.spacingMs,
		Mode:          l.cfg.Mode,
		Window:        l.cfg.Window,
		Limit:         l.cfg.Limit,
		ManualDelayMs: l.cfg.ManualDelayMs,
		Paused:        l.paused,
	}
}

// Reset zeroes the live counters (config is preserved).
func (l *Limiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.totalRequests = 0
	l.delayedCount = 0
	l.passThrough = 0
	l.totalWaitedMs = 0
	l.lastRequestAt = time.Time{}
	l.lastForward = time.Time{}
}

// AvgWaitMs returns the average delay applied across delayed requests, or 0
// when nothing has been delayed yet.
func (s Snapshot) AvgWaitMs() int64 {
	if s.DelayedCount == 0 {
		return 0
	}
	return s.TotalWaitedMs / s.DelayedCount
}
