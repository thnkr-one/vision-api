package progress

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// Bar represents a progress bar display
type Bar struct {
	writer     io.Writer
	width      int
	format     string
	lastUpdate time.Time
	minUpdate  time.Duration
	mu         sync.Mutex
	done       bool
}

// BarStyle defines the visual style of the progress bar
type BarStyle struct {
	LeftBound  string
	RightBound string
	Fill       string
	Empty      string
	Arrow      string
}

// DefaultStyle is the default progress bar style
var DefaultStyle = BarStyle{
	LeftBound:  "[",
	RightBound: "]",
	Fill:       "=",
	Empty:      " ",
	Arrow:      ">",
}

// NewBar creates a new progress bar
func NewBar(writer io.Writer) *Bar {
	if writer == nil {
		writer = io.Discard
	}

	return &Bar{
		writer:     writer,
		width:      30,
		minUpdate:  time.Millisecond * 100, // Prevent too frequent updates
		lastUpdate: time.Now(),
	}
}

// Update updates the progress bar display
func (b *Bar) Update(current, total int64, stats Stats) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.done {
		return
	}

	// Throttle updates
	now := time.Now()
	if now.Sub(b.lastUpdate) < b.minUpdate {
		return
	}
	b.lastUpdate = now

	// Calculate progress
	var percentage float64
	if total > 0 {
		percentage = float64(current) / float64(total) * 100
	}

	// Create progress bar
	completed := int(float64(b.width) * float64(current) / float64(total))
	bar := DefaultStyle.LeftBound
	bar += strings.Repeat(DefaultStyle.Fill, completed)
	if completed < b.width {
		bar += DefaultStyle.Arrow
		bar += strings.Repeat(DefaultStyle.Empty, b.width-completed-1)
	}
	bar += DefaultStyle.RightBound

	// Calculate speed and ETA
	speed := float64(current) / stats.Duration.Seconds()
	var eta time.Duration
	if speed > 0 {
		remainingItems := total - current
		eta = time.Duration(float64(remainingItems)/speed) * time.Second
	}

	// Format the output
	status := fmt.Sprintf("%s %.1f%% | %d/%d",
		bar,
		percentage,
		current,
		total,
	)

	// Add additional stats if available
	if stats.Failed > 0 {
		status += fmt.Sprintf(" | Failed: %d", stats.Failed)
	}
	if stats.Skipped > 0 {
		status += fmt.Sprintf(" | Skipped: %d", stats.Skipped)
	}
	if speed > 0 {
		status += fmt.Sprintf(" | %.1f/s", speed)
	}
	if eta > 0 {
		status += fmt.Sprintf(" | ETA: %s", formatDuration(eta))
	}

	// Clear line and print progress
	fmt.Fprintf(b.writer, "\r\033[K%s", status)
}

// Finish marks the progress bar as complete
func (b *Bar) Finish() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.done {
		fmt.Fprintln(b.writer) // Move to next line
		b.done = true
	}
}

// SetWidth sets the width of the progress bar
func (b *Bar) SetWidth(width int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if width > 0 {
		b.width = width
	}
}

// SetMinUpdateInterval sets the minimum time between updates
func (b *Bar) SetMinUpdateInterval(d time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if d > 0 {
		b.minUpdate = d
	}
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm%ds",
			int(d.Minutes()),
			int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm",
		int(d.Hours()),
		int(d.Minutes())%60)
}

// Stats represents progress statistics
type Stats struct {
	Current   int64
	Total     int64
	Failed    int64
	Skipped   int64
	Duration  time.Duration
	StartTime time.Time
}

// formatBytes formats bytes for display
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(bytes)/float64(div),
		"KMGTPE"[exp])
}
