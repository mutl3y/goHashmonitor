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
	tcfg, err := Config()
	if err != nil {
		t.Fatalf("error configing for test")
	}
	tcfg.Set("Influx.Enabled", true)
	tcfg.Set("Influx.IP", "192.168.0.29")
	tcfg.Set("Influx.DB", "testMine")
	tcfg.Set("Influx.Port", 8086)
	tcfg.Set("influx.User", nil)
	tcfg.Set("Influx.Pw", nil)
	tcfg.Set("Influx.FlushSec", 1*time.Second)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Number of running go routines %v: %v", "before", runtime.NumGoroutine())
			Mine(tcfg)
		})
	}

	time.Sleep(4 * time.Second)
	runtime.Gosched()
	runtime.GC()
	t.Logf("Number of running go routines %v: %v", "after", runtime.NumGoroutine())

	err = pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
	if err != nil {
		t.Logf("pprof didn't work")
	}
	// log.Println(http.ListenAndServe("localhost:6060", nil))
}
