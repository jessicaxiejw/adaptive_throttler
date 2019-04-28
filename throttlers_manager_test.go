package throttlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestIsThrottledAllowsRequestsUptoStartingRatePerHost(t *testing.T) {
	throttles := New(Params{
		StartingRate: 1,
		Burst:        1,
	})

	assert.False(t, throttles.IsThrottled("key"))
	assert.True(t, throttles.IsThrottled("key"))

	assert.False(t, throttles.IsThrottled("new key"))
}

func TestIncrIncrementLimitPerHost(t *testing.T) {
	throttles := New(Params{
		StartingRate: 0,
		Burst:        1,
		Increment:    1,
		UpperBound:   3,
	})

	expectedLimits := []uint64{0, 1, 2, 3, 3, 3}
	for _, expectedLimit := range expectedLimits {
		assert.Equal(t, rate.Limit(expectedLimit), throttles.getThrottler("key").limiter.Limit())
		throttles.Incr("key")
	}
}

func TestDecrDecrementLimitPerHost(t *testing.T) {
	throttles := New(Params{
		StartingRate: 3,
		Burst:        1,
		Decrement:    1,
		LowerBound:   1,
	})

	expectedLimits := []uint64{3, 2, 1, 1, 1}
	for _, expectedLimit := range expectedLimits {
		assert.Equal(t, rate.Limit(expectedLimit), throttles.getThrottler("key").limiter.Limit())
		throttles.Decr("key")
	}
}

func TestCancelWaitPerHost(t *testing.T) {
	// allow no request
	throttles := New(Params{
		StartingRate: 0,
	})

	go throttles.CancelWait("key")
	throttles.Wait("key")
}
