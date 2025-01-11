package progress

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Status represents the current processing status
type Status struct {
	Current   int64
	Total     int64
	Failed    int64
	Skipped   int64
	StartTime time.Time
}

// Tracker handles progress tracking and display
type Tracker struct {
	current   atomic.Int64
	total     atomic.Int64
	failed    atomic.Int64
	skipped   atomic.Int64
	startTime time.Time
	writer    io.Writer
	mu        sync.Mutex
	ticker    *time.Ticker
	done      chan struct{}
}

// NewTracker creates a new progress tracker
func NewTracker(total int64, writer io.Writer) *Tracker {
	if writer == nil {
		writer = os.Stdout
	}

	return &Tracker{
		startTime: time.Now(),
		writer:    writer,
		done:      make(chan struct{}),
		ticker:    time.NewTicker(200 * time.Millisecond),
	}
}

// Start begins progress tracking
func (t *Tracker) Start() {
	t.total.Store(t.total.Load())
	go t.updateDisplay()
}

// Update updates the current progress
func (t *Tracker) Update(current, failed, skipped int64) {
	t.current.Store(current)
	t.failed.Store(failed)
	t.skipped.Store(skipped)
}

// Increment increases the current progress by 1
func (t *Tracker) Increment() {
	t.current.Add(1)
}

// IncrementFailed increases the failed count by 1
func (t *Tracker) IncrementFailed() {
	t.failed.Add(1)
}

// IncrementSkipped increases the skipped count by 1
func (t *Tracker) IncrementSkipped() {
	t.skipped.Add(1)
}

// Finish stops progress tracking
func (t *Tracker) Finish() {
	t.ticker.Stop()
	close(t.done)
	t.displayFinalStatus()
}

func (t *Tracker) updateDisplay() {
	for {
		select {
		case <-t.ticker.C:
			t.displayProgress()
		case <-t.done:
			return
		}
	}
}

func (t *Tracker) displayProgress() {
	t.mu.Lock()
	defer t.mu.Unlock()

	current := t.current.Load()
	total := t.total.Load()
	failed := t.failed.Load()
	skipped := t.skipped.Load()
	elapsed := time.Since(t.startTime)

	// Calculate progress percentage
	var percentage float64
	if total > 0 {
		percentage = float64(current) / float64(total) * 100
	}

	// Calculate speed
	speed := float64(current) / elapsed.Seconds()

	// Create progress bar
	width := 30
	completed := int(float64(width) * float64(current) / float64(total))
	bar := fmt.Sprintf("[%s%s]",
		strings.Repeat("=", completed),
		strings.Repeat(" ", width-completed))

	// Clear line and print progress
	fmt.Fprintf(t.writer, "\r\033[K%s %.1f%% | %d/%d | Failed: %d | Skipped: %d | %.1f/s",
		bar, percentage, current, total, failed, skipped, speed)
}

func (t *Tracker) displayFinalStatus() {
	t.mu.Lock()
	defer t.mu.Unlock()

	current := t.current.Load()
	total := t.total.Load()
	failed := t.failed.Load()
	skipped := t.skipped.Load()
	elapsed := time.Since(t.startTime)

	// Calculate final statistics
	var percentage float64
	if total > 0 {
		percentage = float64(current) / float64(total) * 100
	}
	speed := float64(current) / elapsed.Seconds()

	// Print final status
	fmt.Fprintf(t.writer, "\n\nProcessing completed in %.2f seconds:\n", elapsed.Seconds())
	fmt.Fprintf(t.writer, "  Total files:           %d\n", total)
	fmt.Fprintf(t.writer, "  Successfully processed: %d\n", current)
	fmt.Fprintf(t.writer, "  Failed:                %d\n", failed)
	fmt.Fprintf(t.writer, "  Skipped:               %d\n", skipped)
	fmt.Fprintf(t.writer, "  Success rate:          %.2f%%\n", percentage)
	fmt.Fprintf(t.writer, "  Average speed:         %.2f files/s\n\n", speed)
}

// GetStatus returns the current status
func (t *Tracker) GetStatus() Status {
	return Status{
		Current:   t.current.Load(),
		Total:     t.total.Load(),
		Failed:    t.failed.Load(),
		Skipped:   t.skipped.Load(),
		StartTime: t.startTime,
	}
}