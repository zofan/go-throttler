package throttler

import (
	"encoding/binary"
	"net"
	"sync"
	"time"
)

type Throttler struct {
	Limit  int
	Period time.Duration

	keys map[uint64]key

	mu sync.Mutex
}

type key struct {
	reset     time.Time
	remaining int
}

func New(limit int, period time.Duration) *Throttler {
	return &Throttler{
		Limit:  limit,
		Period: period,
		keys:   make(map[uint64]key),
	}
}

func (t *Throttler) Clean() {
	t.mu.Lock()
	now := time.Now()
	for ip, c := range t.keys {
		if now.Sub(c.reset) < 0 {
			delete(t.keys, ip)
		}
	}
	t.mu.Unlock()
}

func (t *Throttler) Allow(k uint64) (remaining int, reset time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()

	c, ok := t.keys[k]
	if !ok || time.Now().After(c.reset) {
		c = key{remaining: t.Limit, reset: time.Now().Add(t.Period)}
		t.keys[k] = c
	}

	if c.remaining > 0 {
		c.remaining--
	}

	return c.remaining, c.reset
}

func LongIP(ipRaw string) uint64 {
	ip := net.ParseIP(ipRaw)
	if ip == nil {
		return 0
	}

	return binary.BigEndian.Uint64(ip)
}
