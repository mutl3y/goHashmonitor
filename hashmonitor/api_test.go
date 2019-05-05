package hashmonitor

import (
	"fmt"
	_ "net/http"
	_ "net/http/pprof"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"
)

func Test_apiService_Monitor(t *testing.T) {
	t.Logf("Number of running go routines %v: %v", "before", runtime.NumGoroutine())
	tCfg := DefaultConfig()
	tCfg.Set("Influx.DB", "gohashmonitor")
	tCfg.Set("Influx.IP", "192.168.0.29")
	tCfg.Set("Influx.Port", 8086)
	tCfg.Set("Influx.Enabled", true)
	tCfg.Set("Influx.FlushSec", 2*time.Second)

	// if err := ConfigLogger("logging.conf", false); err != nil {
	// }
	fmt.Println(tCfg.GetString(""))
	as := NewStatsService(tCfg).(*apiService)
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		t.Log("recovered from log.panic\n")
	// 	}
	// }()
	met := NewMetricsClient()
	if err := met.Config(tCfg); err != nil {
		log.Infof("failed to config metrics client")
	}

	t.Run("", func(t *testing.T) {

		if ok := as.Monitor(met); ok != true {
			t.Errorf("apiService.Monitor() %v", ok)
		}
		time.Sleep(time.Second * 15)
		if ok := as.stopMonitor(met); ok != true {
			t.Errorf("failed to stop monitor")
		}
	})
	time.Sleep(time.Second * 2)
	t.Logf("Number of running go routines %v: %v", "after", runtime.NumGoroutine())
}

func Test_apiService_StopMonitor(t *testing.T) {
	t.Logf("Number of running go routines %v: %v", "before", runtime.NumGoroutine())
	tcfg := DefaultConfig()
	as := NewStatsService(tcfg).(*apiService)
	met := &metrics{}
	tests := []struct {
		name    string
		api     *apiService
		wantErr bool
	}{
		{"", as, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ok := tt.api.stopMonitor(met); (ok != true) != tt.wantErr {
				t.Errorf("apiService.stopMonitor() error = %v, match %v", ok, tt.wantErr)
			}
		})
	}

	t.Logf("Number of running go routines %v: %v ", "after", runtime.NumGoroutine())
}

func Test_apiService_ShowMonitor(t *testing.T) {
	t.Logf("Number of running go routines %v: %v ", "before", runtime.NumGoroutine())

	c, err := Config()
	if err != nil {
		t.Fatalf("failed to get config")
	}
	ss := NewStatsService(c)
	met := &metrics{}
	ss.Monitor(met)
	// defer ss.stopMonitor()
	time.AfterFunc(10*time.Second, func() {
		ss.stopMonitor(met)
	})

	tests := []struct {
		name    string
		api     ApiService
		wantErr bool
	}{
		{"should work", ss, false},
		// 		{"should break", &apiService{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go tt.api.showMonitor()

		})

	}
	time.Sleep(5 * time.Second)
	t.Logf("Number of running go routines %v: %v ", "after", runtime.NumGoroutine())

}

func Test_apiService_Map(t *testing.T) {
	var testStats stats
	testStats2 := testStats
	threads := []float64{11.1, 12.4, 321.54}
	testStats2.hashrate.Threads = append(testStats2.hashrate.Threads, threads)
	tests := []struct {
		name  string
		stats stats
		want  map[string]interface{}
		match bool
	}{
		{"no threads", testStats, map[string]interface{}{}, true},
		{"thread 0", testStats2, map[string]interface{}{"Thread_0": 11.1}, true},
		{"thread 0 mismatch", testStats2, map[string]interface{}{"Thread_0": 11.2}, false},

		// {"", testStats, map[string]interface{}{}, false},
		// {"", testStats, map[string]interface{}{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stats.Map()
			for mapKey := range tt.want {
				if !reflect.DeepEqual(got[mapKey], tt.want[mapKey]) {
					if tt.match {
						t.Errorf("%v Got  %v %T Want %v %T", mapKey, got[mapKey], got[mapKey], tt.want[mapKey], tt.want[mapKey])
						fmt.Println("local deepequal", DeepEqual(got[mapKey], tt.want[mapKey]))
					}

				}
			}

		})
	}
}

// debug version of reflect.DeepEqual
func DeepEqual(x, y interface{}) bool {
	if x == nil || y == nil {
		return x == y
	}
	v1 := reflect.ValueOf(x)
	v2 := reflect.ValueOf(y)
	if v1.Type() != v2.Type() {
		fmt.Println(v1.Type(), " != ", v2.Type())
		return false
	}
	if v1 != v2 {
		fmt.Println(v1, " != ", v2)
		return false
	}

	return true
}

func Test_simApi(t *testing.T) {

	tcfg, _ := Config()
	t.Run("simApi", func(t *testing.T) {
		api := NewStatsService(tcfg).(*apiService)
		// api.Stats = rwStats{}
		var wg sync.WaitGroup
		wg.Add(1)
		simApi(api, &wg, 4000, 1.5, 500*time.Millisecond)

	})
}
