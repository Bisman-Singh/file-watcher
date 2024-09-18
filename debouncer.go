package main

import (
	"sync"
	"time"
)

// Debouncer coalesces rapid events into a single trigger after a quiet period.
type Debouncer struct {
	interval time.Duration
	timer    *time.Timer
	mu       sync.Mutex
	callback func()
}

// NewDebouncer creates a new Debouncer with the given interval and callback.
func NewDebouncer(interval time.Duration, callback func()) *Debouncer {
	return &Debouncer{
		interval: interval,
		callback: callback,
	}
}

// Trigger resets the debounce timer. The callback fires only after the
// interval has elapsed with no new triggers.
func (d *Debouncer) Trigger() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}

	d.timer = time.AfterFunc(d.interval, d.callback)
}
