package hashmonitor

import (
	"os"
	"runtime"
	"runtime/pprof"

	"testing"
	"time"
)

func Test_mine(t *testing.T) {
	t.Logf("Number of running go routines %v: %v", "before", runtime.NumGoroutine())

	tests := []struct {
		name string
	}{
		{""},
	}

	cfg.Set("Influx.Enabled", true)
	cfg.Set("Influx.IP", "192.168.0.29")
	cfg.Set("influx.DB", "gohashmonitor")
	cfg.Set("Influx.Port", 8086)
	cfg.Set("influx.User", nil)
	cfg.Set("Influx.Pw", nil)
	cfg.Set("Influx.FlushSec", 1*time.Second)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Number of running go routines %v: %v", "before", runtime.NumGoroutine())
			Mine()
		})
	}

	runtime.GC()

	t.Logf("Number of running go routines %v: %v", "after", runtime.NumGoroutine())

	err := pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
	if err != nil {
		t.Logf("pprof didn't work")
	}
	// log.Println(http.ListenAndServe("localhost:6060", nil))
}
