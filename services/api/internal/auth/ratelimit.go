package auth

import (
	"sync"
	"time"
)

// Limiter is a fixed-window rate limiter keyed by an arbitrary string (an
// operation name and client IP, for login and setup). It does not need to
// be clever: a home admin UI only needs to slow down guessing, not resist a
// distributed attacker.
type Limiter struct {
	max    int
	window time.Duration

	mu   sync.Mutex
	hits map[string][]time.Time
}

// NewLimiter allows at most max calls per key within window.
func NewLimiter(max int, window time.Duration) *Limiter {
	return &Limiter{max: max, window: window, hits: make(map[string][]time.Time)}
}

// Allow records an attempt for key and reports whether it is within the
// limit.
func (l *Limiter) Allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := now.Add(-l.window)
	kept := l.hits[key][:0]
	for _, t := range l.hits[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= l.max {
		l.hits[key] = kept
		return false
	}
	l.hits[key] = append(kept, now)
	return true
}
