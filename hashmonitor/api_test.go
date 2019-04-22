package hashmonitor

import (
	"fmt"
	"github.com/spf13/viper"
	"net/http"
	"reflect"
	"runtime"
	"testing"
	"time"
)

import _ "net/http/pprof"
import _ "net/http"

func Test_apiService_Monitor(t *testing.T) {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	t.Logf("Number of running go routines %v: %v", "before", runtime.NumGoroutine())
	tCfg := DefaultConfig()
	tCfg.Set("Influx.DB", "serviceMonitor")
	tCfg.Set("Influx.Enabled", true)
	if err := ConfigLogger("logging.conf", false); err != nil {
	}

	as := NewStatsService(tCfg).(*apiService)

	t.Run("", func(t *testing.T) {
		if ok := as.Monitor(); ok != true {
			t.Errorf("apiService.Monitor() %v", ok)
		}
	})

	time.Sleep(time.Second * 7)
	t.Logf("Number of running go routines %v: %v", "after", runtime.NumGoroutine())
	time.Sleep(time.Second * 4)
	t.Logf("Number of running go routines %v: %v", "after", runtime.NumGoroutine())
}

func Test_apiService_StopMonitor(t *testing.T) {
	t.Logf("Number of running go routines %v: %v", "before", runtime.NumGoroutine())
	cfg := DefaultConfig()
	as := NewStatsService(cfg).(*apiService)

	tests := []struct {
		name    string
		api     *apiService
		wantErr bool
	}{
		{"", as, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ok := tt.api.StopMonitor(); (ok != true) != tt.wantErr {
				t.Errorf("apiService.StopMonitor() error = %v, wantErr %v", ok, tt.wantErr)
			}
		})
	}
	<-as.Signal
	<-as.limit.Signal
	t.Logf("Number of running go routines %v: %v ", "after", runtime.NumGoroutine())
}

func Test_apiService_ShowMonitor(t *testing.T) {
	c, err := Config()
	if err != nil {
		t.Fatalf("failed to get config")
	}
	ss := NewStatsService(c)
	ss.Monitor()
	defer ss.StopMonitor()

	tests := []struct {
		name    string
		api     ApiService
		wantErr bool
	}{
		{"should work", ss, false},
		{"should break", &apiService{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err = tt.api.ShowMonitor(); (err != nil) != tt.wantErr {
				t.Errorf("apiService.ShowMonitor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewAPIService(t *testing.T) {
	type args struct {
		cfg *viper.Viper
	}
	var tests []struct {
		name string
		args args
		want ApiService
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewStatsService(tt.args.cfg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStatsService() = %v, want %v", got, tt.want)
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

func Test_limitClock(t *testing.T) {
	type args struct {
		limit *simpleRateLimit
	}
	var tests []struct {
		name string
		args args
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limitClock(tt.args.limit)
		})
	}
}

func TestNewStatsService(t *testing.T) {
	type args struct {
		cfg *viper.Viper
	}
	var tests []struct {
		name string
		args args
		want ApiService
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewStatsService(tt.args.cfg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStatsService() = %v, want %v", got, tt.want)
			}
		})
	}
}

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

func Test_apiService_Map(t *testing.T) {
	var testStats stats
	testStats2 := testStats
	var blankRes results
	threads := make([]float64, 0, 10)
	threads = append(threads, 11.1, 12.4, 321.54)
	testStats2.hashrate.Threads = append(testStats2.hashrate.Threads, threads)
	tests := []struct {
		name    string
		stats   stats
		want    map[string]interface{}
		wantErr bool
	}{
		{"no threads", testStats, map[string]interface{}{"Connection": connection{}, "Results": blankRes}, false},
		{"thread 0", testStats2, map[string]interface{}{"Connection": connection{}, "Results": results{}, "Threads": []float64{11.1, 12.4, 321.54}}, false},
		{"thread 0 mismatch", testStats2, map[string]interface{}{"Connection": connection{}, "Results": results{}, "Threads": []float64{11.1, 12.4}}, true},

		//{"", testStats, map[string]interface{}{}, false},
		//{"", testStats, map[string]interface{}{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := testStats.Map()
			if (err != nil) != tt.wantErr {
				t.Errorf("apiService.statsToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for mapKey := range got {
				if !reflect.DeepEqual(got[mapKey], tt.want[mapKey]) {
					t.Errorf("apiService.statsToMap() \nGot  %v %T \nwant %v %T", got[mapKey], got[mapKey], tt.want[mapKey], tt.want[mapKey])
					fmt.Println("local deepequal", DeepEqual(got[mapKey], tt.want[mapKey]))

				}
			}
		})
	}
}
