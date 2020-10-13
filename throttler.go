package throttler

import (
	"sync"
	"time"
)

type Throttler struct {
	limit  int
	period time.Duration

	keys map[uint64]*key

	mu sync.Mutex
}

type key struct {
	reset     time.Time
	remaining int
}

func New(limit int, period time.Duration) *Throttler {
	return &Throttler{
		limit:  limit,
		period: period,
		keys:   make(map[uint64]*key),
	}
}

func (t *Throttler) Limit() int {
	return t.limit
}

func (t *Throttler) Period() time.Duration {
	return t.period
}

func (t *Throttler) Clean() {
	t.mu.Lock()
	now := time.Now()
	for k, c := range t.keys {
		if now.After(c.reset) {
			delete(t.keys, k)
		}
	}
	t.mu.Unlock()
}

func (t *Throttler) Allow(k uint64) (remaining int, reset time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	_, ok := t.keys[k]
	if !ok || now.After(t.keys[k].reset) {
		t.keys[k] = &key{remaining: t.limit, reset: now.Add(t.period)}
	}

	if t.keys[k].remaining > 0 {
		t.keys[k].remaining--
	}

	return t.keys[k].remaining, t.keys[k].reset
}
