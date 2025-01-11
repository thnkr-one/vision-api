package rate

import (
	"context"
	"sync"
	"time"
)

// Limiter provides rate limiting functionality with a sliding window
type Limiter struct {
	mu          sync.Mutex
	rate        int           // Maximum requests per window
	window      time.Duration // Time window for rate limiting
	requests    []time.Time   // Sliding window of request timestamps
	maxWaitTime time.Duration // Maximum time to wait for a token
}

// NewLimiter creates a new rate limiter
// rate: maximum number of requests
// window: time window for rate limiting
func NewLimiter(rate int, window time.Duration) *Limiter {
	return &Limiter{
		rate:        rate,
		window:      window,
		requests:    make([]time.Time, 0, rate),
		maxWaitTime: window,
	}
}

// Wait blocks until a request can be made or context is canceled
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		if delay, allow := l.tryAcquire(); allow {
			return nil
		} else {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Continue trying
			}
		}
	}
}

// tryAcquire attempts to acquire a token
// Returns the delay to wait if not allowed and whether the request is allowed
func (l *Limiter) tryAcquire() (time.Duration, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	// Remove expired timestamps
	valid := 0
	for _, t := range l.requests {
		if t.After(cutoff) {
			l.requests[valid] = t
			valid++
		}
	}
	l.requests = l.requests[:valid]

	// Check if we can make a request
	if len(l.requests) < l.rate {
		l.requests = append(l.requests, now)
		return 0, true
	}

	// Calculate delay for next available slot
	nextSlot := l.requests[0].Add(l.window)
	delay := nextSlot.Sub(now)

	if delay > l.maxWaitTime {
		return delay, false
	}

	return delay, false
}

// SetMaxWaitTime sets the maximum time to wait for a token
func (l *Limiter) SetMaxWaitTime(d time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.maxWaitTime = d
}

// GetCurrentRate returns the current rate of requests
func (l *Limiter) GetCurrentRate() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	// Count only non-expired requests
	count := 0
	for _, t := range l.requests {
		if t.After(cutoff) {
			count++
		}
	}

	return count
}

// Reset clears all stored request timestamps
func (l *Limiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.requests = l.requests[:0]
}

// Available returns the number of requests available in the current window
func (l *Limiter) Available() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	// Count only non-expired requests
	count := 0
	for _, t := range l.requests {
		if t.After(cutoff) {
			count++
		}
	}

	return l.rate - count
}