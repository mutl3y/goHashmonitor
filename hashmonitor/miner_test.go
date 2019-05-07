package hashmonitor

import (
	"io"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestNewMiner(t *testing.T) {
	m := NewMiner()
	tests := []struct {
		name string
		want *miner
	}{
		{"", m},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMiner(); !reflect.DeepEqual(got.config, tt.want.config) {
				t.Errorf("NewMiner() \n Got %v \nwant %v", *got, *tt.want)

			}
		})
	}
}

func Test_ConfigMiner(t *testing.T) {
	c, err := Config()
	if err != nil {
		t.Fatalf("Failed to get config %v", err)
	}
	m := NewMiner()

	tests := []struct {
		name         string
		ms           *miner
		cfg          *viper.Viper
		wantError    bool
		errorMessage string
	}{
		{"Should Fail", m, &viper.Viper{}, true, "stak Directory Not Specified"},
		{"Should Work", m, c, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.ms.ConfigMiner(tt.cfg)
			if err != nil {
				if err.Error() != tt.errorMessage {
					t.Fatalf("%v\n want: %v\n got: %v ", tt.name, tt.errorMessage, err)
				}
			}

			// 	t.Logf("want: %v \ngot %v", cfg, tt.ms.config)
		})
	}
}

func TestMiner_StartMining_StopMining(t *testing.T) {
	c, err := Config()
	if err != nil {
		t.Fatalf("Failed to get config %v", err)
	}

	m := NewMiner()
	ctx, err := m.ConfigMiner(c)
	if err != nil {
		t.Fatalf("Failed configuring miner: %v", err)
	}
	err = m.StartMining(ctx)
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(3 * time.Second)
	err = m.StopMining()
	if err != nil {
		t.Fatalf("failed to stop mining process %v", err)
	}

}

func TestMiner_ConsoleMetrics(t *testing.T) {
	tCfg := viper.New()
	// if err != nil {
	// 	t.Fatalf("Failed to get config %v", err)
	// }
	tCfg.Set("Influx.Enabled", true)
	tCfg.Set("Influx.DB", "consoleMetrics")
	tCfg.Set("Influx.Retention", time.Second*10)
	tCfg.Set("Influx.Ip", "192.168.0.29")
	tCfg.Set("Influx.Port", 8086)
	tCfg.Set("Core.Stak.Dir", root+"xmr-stak")
	tCfg.Set("Core.Stak.Exe", "./xmr-stak.exe")

	m := NewMiner()
	_, err := m.ConfigMiner(tCfg)
	if err != nil {
		t.Fatalf("Failed configuring miner: %v", err)
	}

	f, err := os.Open(".testcode" + pathSep + "stakMiningOutput")
	if err != nil {
		t.Fatalf("open file error %v", err)
	}
	//noinspection ALL
	defer f.Close()

	met := NewMetricsClient()
	if err = met.Config(tCfg); err != nil {
		debug("error %v", err)
	}

	go met.backGroundWriter()

	testData := io.ReadCloser(f)
	m.StdOutPipe = &testData
	m.ConsoleMetrics(met)
	met.Stop()
}

func TestInterleaveFilter(t *testing.T) {
	// 	d := int64(155332121)
	if err := ConfigLogger("logging.conf", false); err != nil {
		t.Fatal("failed configuring logger")
	}

	tests := []struct {
		name    string
		msg     string
		wantErr bool
	}{
		{"Works", "99|66: 73/1983.20 ms - 2", false},
		{"nan", "9a9|66: 73/1983.20 ms - 2", true},
		{"nan2", "99|66: 73/1a983.20 ms - 2", true},
		{"n/a", "N/A ", true},
		{"", "[2019-05-06 22:03:21] : OpenCL 0|1: auto-tune validate intensity 520|512", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := interleaveFilter(tt.msg); (err != nil) != tt.wantErr {
				t.Errorf("interleaveFilter() error = %v, match %v", err, tt.wantErr)
			}
		})
	}
}
func TestAutotuneFilter(t *testing.T) {
	// 	d := int64(155332121)
	if err := ConfigLogger("logging.conf", false); err != nil {
		t.Fatal("failed configuring logger")
	}

	tests := []struct {
		name    string
		msg     string
		wantErr bool
	}{
		{"", "OpenCL 0|1: auto-tune validate intensity 520|512", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := autotuneFilter(tt.msg); (err != nil) != tt.wantErr {
				t.Errorf("interleaveFilter() error = %v, match %v", err, tt.wantErr)
			}
		})
	}
}
