package throttlers

import "sync"

// Manager is thread-safe. It handles the rate limit information of multiple keys.
type Manager struct {
	throttlers   sync.Map
	startingRate uint64 // all numbers are in request per second
	burst        uint64
	lowerBound   uint64
	upperBound   uint64
	increment    uint64
	decrement    uint64
}

// Params is used for setting up a new Manager in the New() method. All fields in the Params are in request per second.
type Params struct {
	// StartingRate is the default request rate per second which Manager permits. If the value is set to 0, it will never send any request. It is highly recommend that you do NOT set StartingRate to 0.
	StartingRate uint64

	// The maximum burst. It allows more events happen at once.
	Burst uint64

	// LowerBound is the lowest request rate per second for a key.
	// If the value is set to 0, once the request rate reaches LowerBound. It won't send any more requests until a restart.
	LowerBound uint64

	// UpperBound is the highest request rate per second for a key.
	UpperBound uint64

	// Everytime the Incr(key) function is called, the manager increases the current request rate for a given key by this value.
	Increment uint64

	// Everytime the Decr(key) function is called, the manager decreases the current request rate for a given key by this value.
	Decrement uint64
}

// New creates a new manager that stores the rate limit for multiple keys.
func New(params Params) *Manager {
	return &Manager{
		startingRate: params.StartingRate,
		burst:        params.Burst,
		lowerBound:   params.LowerBound,
		upperBound:   params.UpperBound,
		increment:    params.Increment,
		decrement:    params.Decrement,
	}
}

// Wait sleeps until the request for the key can be executed (i.e. no longer throttled).
// Wait is the recommended method by golang.org/x/time/rate
func (t *Manager) Wait(key string) {
	t.getThrottler(key).Wait()
}

// CancelWait stops the current wait for the key. Once it is called, all subsequent calls do nothing.
// It is usually called during a shut down.
func (t *Manager) CancelWait(key string) {
	t.getThrottler(key).CancelWait()
}

// IsThrottled checks if the current key is throttled. Calling the method counts towards its throttled limit.
// For example, if "foo" key only allows 2 req/s, calling IsThrottled("foo") used up one request.
// Thus, it allows 1 req/s until the 1 second window is over.
func (t *Manager) IsThrottled(key string) bool {
	return t.getThrottler(key).IsThrottled()
}

// Incr increases the current throttle limit by the Increment amount set in the New method.
// People usually call this method after a successful request.
func (t *Manager) Incr(key string) {
	t.getThrottler(key).Incr()
}

// Decr decreases the current throttle limit by the Decrement amount set in the New method.
// People usually call this method after a failing request.
func (t *Manager) Decr(key string) {
	t.getThrottler(key).Decr()
}

func (t *Manager) getThrottler(key string) *throttler {
	if th, ok := t.throttlers.Load(key); ok {
		return th.(*throttler)
	}

	th, _ := t.throttlers.LoadOrStore(key, newThrottler(paramThrottler{
		StartingRate: t.startingRate,
		Burst:        t.burst,
		LowerBound:   t.lowerBound,
		UpperBound:   t.upperBound,
		Increment:    t.increment,
		Decrement:    t.decrement,
	}))
	return th.(*throttler)
}
