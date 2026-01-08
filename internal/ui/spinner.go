package ui

import (
	"fmt"
	"time"
)

// Spinner represents a simple CLI spinner
type Spinner struct {
	stopChan chan struct{}
	message  string
}

// NewSpinner creates a new spinner with a message
func NewSpinner(message string) *Spinner {
	return &Spinner{
		stopChan: make(chan struct{}),
		message:  message,
	}
}

// Start starts the spinner in a background goroutine
func (s *Spinner) Start() {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		i := 0
		for {
			select {
			case <-s.stopChan:
				return
			case <-ticker.C:
				fmt.Printf("\r%s%s %s%s", ColorCyan, frames[i%len(frames)], ColorReset, s.message)
				i++
			}
		}
	}()
}

// Stop stops the spinner and clears the line
func (s *Spinner) Stop() {
	close(s.stopChan)
	fmt.Print("\r\033[K") // Clear the line
}
