package hashmonitor

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func Test_apiService_Monitor(t *testing.T) {
	t.Logf("Number of running go routines %v: %v", "before", runtime.NumGoroutine())
	cfg := DefaultConfig()
	as := NewStatsService(cfg).(*apiService)

	tests := []struct {
		name    string
		api     *apiService
		wantErr bool
	}{
		{"New ApiService", as, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ok := tt.api.Monitor(); (ok != true) != tt.wantErr {
				t.Errorf("apiService.Monitor() error = %v, wantErr %v", ok, tt.wantErr)
			}
		})
	}

	fmt.Printf("%v\n", as.StopMonitor())
	time.Sleep(time.Second * 2)
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
