# Adaptive Throttler

A thread-safe throttler library that
- handles multiple throttlers within one manager
- allows you ramp up and ramp down request rate per throttler
- generic for different use cases(eg/ controls the rate limit per host, limit database reads/writes, and etc)

The adaptive throttler is a wrapper around the [golang.org/x/time/rate](https://godoc.org/golang.org/x/time/rate) library.

## Installation
`go get github.com/jessicaxiejw/adaptive_throttler`

See [godoc](https://godoc.org/github.com/jessicaxiejw/adaptive_throttler) for in-depth explanation on the functions and parameters.

## Example Usage

```golang
manager := throttlers.New(throttlers.Params{
	StartingRate: 10,	// request per second
	Burst: 5,
	LowerBound: 1,
	UpperBound: 20,
	Increment: 2,
	Decrement: 1,
})

key := "http://example.com/"
for {
	manager.Wait(key)	// waiting until request can be sent

	resp, err := http.Get(key)
	if err != nil {
		manager.Decrement(key)
	} else {
		manager.Increment(key)
	}
}
```

# How does the package work?
When the throttlers is first created, it will use `StartingRate` for every key. The request rate per key is adjusted based on the `Increment` and `Decrement` call. For example, say we set

```
StartingRate: 10
Burst: 1
LowerBound: 1
UpperBound: 20
Increment: 2
Decrement: 1
```

We would at first allow 10 requests/s for a given key. Two requests were successfully and we called the `Increment` function twice. The request rate limit was raised to StartingRate + 2 * Increment = 10 + 2 * 2 = 14 request/s.

Right after, one request wasn't successful and you called `Decrement`. The rate limit was now set to (current request rate) - Decrement = 14 - 1 = 13 requests/s.

We then successfully sent 4 requests, the new request rate should be (current request rate) + Increment * 4 = 13 + 2 * 4 = 21 requests/s. However, because the UpperBound was set to 20 requests/s, the new request rate actually became 20 requests/s.

Say there were 20 unsuccessful requests in series, the new request rate would hit the LowerBound, which was 1 requests/s.

# Can the package only be used for handling HTTP requests?
No.

Even though the package was originally created for keeping track of throttle limits for different hosts. It is designed to be generic for any types of throttling. For example, you can use it for rate limiting a database's read and write per user. The key will be the user name. 
