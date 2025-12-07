package limiter

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
)

type Limiter struct {
	semaphore   chan struct{}
	rateLimiter *rate.Limiter
	mu          sync.Mutex
}

func New(maxConcurrent, ratePerSecond int) *Limiter {
	return &Limiter{
		semaphore:   make(chan struct{}, maxConcurrent),
		rateLimiter: rate.NewLimiter(rate.Limit(ratePerSecond), ratePerSecond),
	}
}

func (l *Limiter) Acquire(ctx context.Context) (release func(), err error) {
	if err := l.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	select {
	case l.semaphore <- struct{}{}:
		return func() { <-l.semaphore }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (l *Limiter) TryAcquire() (release func(), ok bool) {
	if !l.rateLimiter.Allow() {
		return nil, false
	}

	select {
	case l.semaphore <- struct{}{}:
		return func() { <-l.semaphore }, true
	default:
		return nil, false
	}
}
