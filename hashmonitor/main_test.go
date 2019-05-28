package hashmonitor

import (
	"github.com/spf13/viper"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMine(t *testing.T) {
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
	defer api.stopMonitor(met)

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
		s MineSession
		c *viper.Viper
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"int 100 fail ", args{MineSession{}, tcfg}, true},
		{"int 100", args{MineSession{
			"xmr-stak/amd.txt", api, cards, met, amdConf}, tcfg}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			time.AfterFunc(30*time.Second, func() { tt.args.s.stopMine() })
			err = tt.args.s.Mine(tt.args.c)
			if (err != nil) && !tt.wantErr {
				t.Fatalf("%v", err)
			}
		})
	}

}
