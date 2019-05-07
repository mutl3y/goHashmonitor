package hashmonitor

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func Test_stats_startingHash(t *testing.T) {

	tests := []struct {
		name       string
		uptime     int
		stableTime time.Duration
		wantErr    bool
	}{
		{"nil", 0, time.Millisecond, false},
		{"uptime 2 stable 0", 2, time.Second, false},
		{"uptime 2 stable 2", 2, time.Second * 2, false},
		{"uptime 2 stable 4", 2, time.Second * 4, false},
		{"uptime 4 stable 2", 4, time.Second * 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &apiService{}
			s.Stats.data.Threads = [][]float64{{10000.0}}
			s.Stats.data.Uptime = tt.uptime
			if err := s.startingHash(500, tt.stableTime); (err != nil) != tt.wantErr {
				t.Fatalf("%v", err)
			}

		})
	}
}

func Test_stats_minHash(t *testing.T) {
	t.Logf("Number of running go routines %v: %v", "before", runtime.NumGoroutine())
	tCfg := DefaultConfig()
	tCfg.Set("Influx.DB", "serviceMonitor")
	tCfg.Set("Influx.Enabled", true)

	tests := []struct {
		name          string
		min, hashrate int
		wantErr       bool
	}{
		{"min 500 hashrate nil", 500, 0, true},
		{"min 500 hashrate 100", 500, 100, true},
		{"min 500 hashrate 500", 500, 500, false},
		{"min 500 hashrate 1500", 500, 1500, false},
		{"min 500 hashrate 9900", 5000, 9900, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &apiService{}
			api.Stats.data.Total = []float64{float64(tt.hashrate)}
			if err := api.minHash(tt.min); (err != nil) != tt.wantErr {
				t.Errorf("stats.minHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_stats_currentHash(t *testing.T) {
	tCfg := DefaultConfig()
	tCfg.Set("Influx.DB", "serviceMonitor")
	tCfg.Set("Influx.Enabled", true)
	api := NewStatsService(tCfg).(*apiService)
	tests := []struct {
		name                      string
		hash, hashrate, maxErrors int
		refresh                   time.Duration
		decayRate                 float64
		wantErr                   bool
	}{
		// {"", 2000, 4000, 3, time.Second, 1, false},
		{"", 2000, 4000, 3, time.Second, 1.25, true},
	}

	for _, tt := range tests {
		stat := stats{}
		stat.Total = []float64{float64(tt.hashrate)}
		debug("tch %v", *api)
		api.StatsUpdate(stat)
		var wg sync.WaitGroup
		wg.Add(1)
		go simApi(api, &wg, tt.hashrate, tt.decayRate, 500*time.Millisecond)
		wg.Wait()
		t.Run(tt.name, func(t *testing.T) {

			// api.Stats = rwStats{}

			if err := api.currentHash(tt.hash, tt.maxErrors, tt.refresh); (err != nil) != tt.wantErr {
				t.Errorf("stats.currentHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
