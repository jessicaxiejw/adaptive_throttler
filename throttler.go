package throttlers

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
)

type throttler struct {
	limiter *rate.Limiter

	lowerBound   uint64 // in rps
	upperBound   uint64
	increment    uint64
	decrement    uint64
	currentLimit uint64

	ctx    context.Context
	cancel context.CancelFunc

	mutex sync.Mutex
}

type paramThrottler struct {
	StartingRate, Burst, LowerBound, UpperBound, Increment, Decrement uint64
}

func newThrottler(param paramThrottler) *throttler {
	ctx, cancel := context.WithCancel(context.Background())

	return &throttler{
		limiter:      rate.NewLimiter(rate.Limit(param.StartingRate), int(param.Burst)),
		lowerBound:   param.LowerBound,
		upperBound:   param.UpperBound,
		increment:    param.Increment,
		decrement:    param.Decrement,
		currentLimit: param.StartingRate,
		ctx:          ctx,
		cancel:       cancel,
	}
}

func (t *throttler) Wait() {
	t.limiter.Wait(t.ctx)
}

func (t *throttler) CancelWait() {
	t.cancel()
}

func (t *throttler) IsThrottled() bool {
	return !t.limiter.Allow()
}

func (t *throttler) Incr() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.currentLimit == t.upperBound {
		return
	}

	t.currentLimit += t.increment
	if t.currentLimit >= t.upperBound {
		t.currentLimit = t.upperBound
	}

	t.limiter.SetLimit(rate.Limit(t.currentLimit))
}

func (t *throttler) Decr() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.currentLimit == t.lowerBound {
		return
	}

	t.currentLimit -= t.decrement
	if t.currentLimit <= t.decrement {
		t.currentLimit = t.lowerBound
	}

	t.limiter.SetLimit(rate.Limit(t.currentLimit))
}
