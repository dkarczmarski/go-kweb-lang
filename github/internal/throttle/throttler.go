package throttle

import (
	"context"
	"sync"
	"time"
)

type Throttler struct {
	mu       sync.Mutex
	last     time.Time
	interval time.Duration
}

func NewThrottler(interval time.Duration) *Throttler {
	return &Throttler{interval: interval}
}

func (t *Throttler) Throttle(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.last.IsZero() {
		elapsed := time.Since(t.last)
		if remaining := t.interval - elapsed; remaining > 0 {
			if err := sleepCtx(ctx, remaining); err != nil {
				return err
			}
		}
	}

	t.last = time.Now()
	return nil
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
