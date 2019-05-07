package hashmonitor

import (
	"runtime"
	"time"
)

type simpleRateLimit struct {
	Signal   chan bool
	throttle chan time.Time
	rate     time.Duration
}

func limitClock(limit *simpleRateLimit) {
	// pc uintptr, file string, line int, ok bool
	var (
		file string
		line int
	)
	_, file, line, _ = runtime.Caller(2)

	debug("Starting Rate Limiter %v %v", file, line)

	debug(GetCallerFunctionName())
	debug(GetCurrentFunctionName())

	tick := time.NewTicker(limit.rate)
	// limiter:

	for t := range tick.C {
		select {
		case <-limit.Signal:
			// close(limit.throttle)
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

func GetCurrentFunctionName() string {
	// Skip GetCurrentFunctionName
	return getFrame(1).Function
}

func GetCallerFunctionName() string {
	// Skip GetCallerFunctionName and the function to get the caller of
	return getFrame(2).Function
}

func getFrame(skipFrames int) runtime.Frame {
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := skipFrames + 2

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}
