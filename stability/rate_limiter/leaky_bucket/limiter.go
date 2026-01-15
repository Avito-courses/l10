package main

import (
	"sync"
	"time"
)

type LeakyBucket struct {
	capacity int
	rate     int
	queue    []time.Time
	mu       sync.Mutex
}

func NewLeakyBucket(capacity, rate int) *LeakyBucket {
	lb := &LeakyBucket{
		capacity: capacity,
		rate:     rate,
		queue:    make([]time.Time, 0, capacity),
	}
	go lb.leak()
	return lb
}

func (lb *LeakyBucket) Allow() bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if len(lb.queue) < lb.capacity {
		lb.queue = append(lb.queue, time.Now())
		return true
	}

	return false
}

func (lb *LeakyBucket) leak() {
	ticker := time.NewTicker(time.Second / time.Duration(lb.rate))
	defer ticker.Stop()

	for range ticker.C {
		lb.mu.Lock()
		if len(lb.queue) > 0 {
			lb.queue = lb.queue[1:]
		}
		lb.mu.Unlock()
	}
}
