package throttlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestIsThrottledAllowsRequestsUptoStartingRate(t *testing.T) {
	throttle := newThrottler(paramThrottler{
		StartingRate: 1,
		Burst:        1,
	})

	assert.False(t, throttle.IsThrottled())
	assert.True(t, throttle.IsThrottled())
}

func TestIncrIncrementLimit(t *testing.T) {
	throttle := newThrottler(paramThrottler{
		StartingRate: 0,
		Burst:        1,
		Increment:    1,
		UpperBound:   3,
	})

	expectedLimits := []uint64{0, 1, 2, 3, 3, 3}
	for _, expectedLimit := range expectedLimits {
		assert.Equal(t, rate.Limit(expectedLimit), throttle.limiter.Limit())
		throttle.Incr()
	}
}

func TestDecrDecrementLimit(t *testing.T) {
	throttle := newThrottler(paramThrottler{
		StartingRate: 3,
		Burst:        1,
		Decrement:    1,
		LowerBound:   1,
	})

	expectedLimits := []uint64{3, 2, 1, 1, 1}
	for _, expectedLimit := range expectedLimits {
		assert.Equal(t, rate.Limit(expectedLimit), throttle.limiter.Limit())
		throttle.Decr()
	}
}

func TestCancelWait(t *testing.T) {
	// allow no request
	throttle := newThrottler(paramThrottler{
		StartingRate: 0,
	})

	go throttle.CancelWait()
	throttle.Wait()
}
