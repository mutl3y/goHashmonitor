package hashmonitor

import (
	"time"
)

type simpleRateLimit struct {
	Signal   chan bool
	throttle chan time.Time
	rate     time.Duration
}

func limitClock(limit *simpleRateLimit) {
	debug("starting Rate Limiter")
	tick := time.NewTicker(limit.rate)

	for t := range tick.C {
		select {
		case <-limit.Signal:
			return
		case limit.throttle <- t:
		}
	}

	debug("Stopped Rate Limiter")
}

func (limit *simpleRateLimit) Stop() bool {
	limit.Signal <- true
	debug("Stopping limit Service")
	return true
}

// newLimiter Takes a time duration for refresh speed
// returns a simple rate limiter Config with a signal channel
// uses select so non blocking
func newLimiter(rate time.Duration) *simpleRateLimit {
	t := make(chan time.Time, 1)
	c := make(chan bool)
	limit := simpleRateLimit{Signal: c, throttle: t, rate: rate}
	return &limit
}
