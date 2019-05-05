package hashmonitor

import (
	"reflect"
	"runtime"
	"testing"
	"time"
)

func Test_limitClock(t *testing.T) {
	debug("test")
	lim := newLimiter(1 * time.Second)
	tests := []struct {
		name  string
		limit *simpleRateLimit
	}{
		{"", lim},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Number of running go routines %v: %v", "before", runtime.NumGoroutine())
			go limitClock(tt.limit)
			<-tt.limit.throttle
			tt.limit.Stop()
			time.Sleep(1 * time.Second)
			t.Logf("Number of running go routines %v: %v", "after", runtime.NumGoroutine())

		})
	}
}

func Test_simpleRateLimit_Stop(t *testing.T) {
	type fields struct {
		Signal   chan bool
		throttle chan time.Time
		rate     time.Duration
	}
	var tests []struct {
		name   string
		fields fields
		want   bool
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit := &simpleRateLimit{
				Signal:   tt.fields.Signal,
				throttle: tt.fields.throttle,
				rate:     tt.fields.rate,
			}
			if got := limit.Stop(); got != tt.want {
				t.Errorf("simpleRateLimit.Stop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newLimiter(t *testing.T) {
	type args struct {
		rate time.Duration
	}
	var tests []struct {
		name string
		args args
		want *simpleRateLimit
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newLimiter(tt.args.rate); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newLimiter() = %v, want %v", got, tt.want)
			}
		})
	}
}
