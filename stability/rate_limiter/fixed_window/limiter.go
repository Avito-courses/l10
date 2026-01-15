package main

import (
	"sync"
	"time"
)

type FixedWindow struct {
	limit       int
	window      time.Duration
	counter     int
	windowStart time.Time
	mu          sync.Mutex
}

func NewFixedWindow(limit int, window time.Duration) *FixedWindow {
	return &FixedWindow{
		limit:       limit,
		window:      window,
		counter:     0,
		windowStart: time.Now(),
	}
}

func (fw *FixedWindow) Allow() bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	now := time.Now()

	if now.Sub(fw.windowStart) >= fw.window {
		fw.counter = 0
		fw.windowStart = now
	}

	if fw.counter < fw.limit {
		fw.counter++
		return true
	}

	return false
}
