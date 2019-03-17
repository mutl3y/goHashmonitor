package hashmonitor

import (
	"fmt"
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
			cfg, err := tt.ms.ConfigMiner(tt.cfg)
			if err != nil {
				if err.Error() != tt.errorMessage {
					t.Fatalf("%v\n want: %v\n got: %v ", tt.name, tt.errorMessage, err)
				}
			}

			t.Logf("want: %v \ngot %v", cfg, tt.ms.config)
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
		fmt.Printf("%v", err)
	}
	time.Sleep(3 * time.Second)
	err = m.StopMining()
	if err != nil {
		t.Fatalf("failed to stop mining process %v", err)
	}

}

func TestMiner_ConsoleMetrics(t *testing.T) {
	c, err := Config()
	if err != nil {
		t.Fatalf("Failed to get config %v", err)
	}

	m := NewMiner()

	_, err = m.ConfigMiner(c)
	if err != nil {
		t.Fatalf("Failed configuring miner: %v", err)
	}

	f, err := os.Open(".testcode" + pathSep + "stakMiningOutput")
	if err != nil {
		t.Fatalf("open file error %v", err)
	}
	//noinspection ALL
	defer f.Close()
	testData := io.ReadCloser(f)
	m.StdOutPipe = &testData
	if err := m.ConsoleMetrics(); err != nil {
		t.Errorf("miner.ConsoleMetrics() %v", err)
	}

}

func TestInterleaveFilter(t *testing.T) {
	d := int64(155332121)

	tests := []struct {
		name    string
		msg     string
		wantErr bool
	}{
		{"Works", "99|66: 73/1983.20 ms - 2", false},
		{"nan", "9a9|66: 73/1983.20 ms - 2", true},
		{"", "99|66: 73/1a983.20 ms - 2", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := interleaveFilter(tt.msg, d); (err != nil) != tt.wantErr {
				t.Errorf("interleaveFilter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
