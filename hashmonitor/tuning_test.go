package hashmonitor

import (
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func Test_TuningRun(t *testing.T) {
	t.Logf("Number of running go routines %v: %v", "before", runtime.NumGoroutine())

	tests := []struct {
		name string
		args IntensityRun
	}{
		// {"reset Disabled, 10 second runtime", IntensityRun{
		// 	Intensity:  IntRunArgs{Start: 100, Stop: 100, Inc: 1},
		// 	Worksize:   IntRunArgs{Start: 2, Stop: 3, Inc: 1},
		// 	Runtime:    10 * time.Second,
		// 	ResetCards: false},
		// },
		// {"reset Disabled, 10 second runtime", IntensityRun{
		// 	Intensity:  IntRunArgs{Start: 100, Stop: 120, Inc: 10},
		// 	Worksize:   IntRunArgs{Start: 2, Stop: 2, Inc: 1},
		// 	Runtime:    10 * time.Second,
		// 	ResetCards: false},
		// },
		{"reset Disabled, 10 second runtime", IntensityRun{
			Intensity: IntRunArgs{Start: 800, Stop: 1900, Inc: 300},
			Worksize:  IntRunArgs{Start: 8, Stop: 20, Inc: 2},
			Runtime:   1 * time.Minute,
			AutoTune:  4, Interleave: 20, AfterAllLock: 10 * time.Second, ResetCards: false},
		},
	}
	tcfg, err := Config()
	if err != nil {
		t.Fatalf("error configing for test")
	}
	tcfg.Set("Influx.Enabled", true)
	tcfg.Set("Influx.IP", "192.168.0.29")
	tcfg.Set("Influx.DB", "tryme")
	tcfg.Set("Influx.Port", 8086)
	tcfg.Set("influx.User", nil)
	tcfg.Set("Influx.Pw", nil)
	tcfg.Set("Influx.FlushSec", 1*time.Second)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err = TuningRun(tcfg, tt.args)
			if err != nil {
				t.Fatalf("TuningRun %v", err)
			}
		})
	}
	t.Logf("Number of running go routines %v: %v", "after", runtime.NumGoroutine())

	time.Sleep(10 * time.Second)
	// runtime.Gosched()
	// runtime.GC()
	t.Logf("Number of running go routines %v: %v", "after", runtime.NumGoroutine())

	err = pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
	if err != nil {
		t.Logf("pprof didn't work")
	}
}

func TestRunMiner(t *testing.T) {
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

	cards := NewCardData(tcfg)
	cards.resetEnabled = false
	err = cards.GetStatus()
	if err != nil {
		t.Fatalf("%+v\n", err)
	}

	if err = cards.ResetCards(false); err != nil {
		t.Fatalf("error Resetting cards %v\n", err)
	}

	api := NewStatsService(tcfg).(*apiService)

	met := NewMetricsClient()

	met.enabled = true
	met.refresh = 10 * time.Second
	met.db = "tuningRun"
	if err = met.Config(tcfg); err != nil {
		log.Infof("failed to config metrics client")
	}

	go met.backGroundWriter()

	go api.Monitor(met)
	dir := tcfg.GetString("Core.Stak.Dir")
	if dir == "" {
		dir = "xmr-stak"
	}

	file := strings.Join([]string{dir, "amd.txt"}, pathSep)
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("Can't find File %v, %v", file, err)
	}

	amdConf := NewAmdConfig()
	if err = amdConf.gpuConfParse(f); err != nil {
		log.Errorf("AmdConf.gpuConfParse() error = %v", err)
	}
	f.Close()
	type args struct {
		s tuneSession
		c *viper.Viper
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"int 100", args{tuneSession{}, tcfg}, true},
		{"int 100", args{tuneSession{
			"xmr-stak/amd.txt",
			api, cards, met, amdConf,
			1000, 18, 20, 0,
			10 * time.Second, 3 * time.Second}, tcfg}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = RunMiner(tt.args.s, tt.args.c)
			if (err != nil) && !tt.wantErr {
				t.Fatalf("%v", err)
			}
		})
	}
	api.stopMonitor(met)
}

func Test_highestDiv(t *testing.T) {
	type args struct {
		numerator int
		divider   int
	}
	tests := []struct {
		name  string
		args  args
		want  int
		match bool
	}{
		{"1", args{100, 10}, 100, true},
		{"2", args{100, 9}, 99, true},
		{"3", args{100, 6}, 96, true},
		{"4", args{100, 6}, 97, false},
		{"5", args{100, 101}, 101, true},
		{"6", args{100, 15}, 90, true},
		{"7", args{103, 3}, 102, true},
		{"8", args{1992, 16}, 1984, true},
		{"9", args{10, 16}, 16, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := highestDiv(tt.args.numerator, tt.args.divider)

			if (got != tt.want) == tt.match {
				t.Errorf("highestDiv() %v got %v, want %v", tt.name, got, tt.want)
			} else {
				t.Logf("wannt %v got %v", tt.want, got)
			}
		})
	}
}
func Test_InterleaveRun(t *testing.T) {
	t.Logf("Number of running go routines %v: %v", "before", runtime.NumGoroutine())

	tests := []struct {
		name string
		args InterleaveRun
	}{

		{"reset Disabled, 10 second runtime", InterleaveRun{
			Interleave: IntRunArgs{Start: 20, Stop: 40, Inc: 5},
			Runtime:    90 * time.Second,
			ResetCards: false},
		},
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

			err = InterleaveSession(tcfg, tt.args)
			if err != nil {
				t.Fatalf("TuningRun %v", err)
			}
		})
	}
	t.Logf("Number of running go routines %v: %v", "after", runtime.NumGoroutine())

	t.Logf("Number of running go routines %v: %v", "after", runtime.NumGoroutine())

	err = pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
	if err != nil {
		t.Logf("pprof didn't work")
	}

}
