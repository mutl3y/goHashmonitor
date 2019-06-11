package hashmonitor

import (
	"os"
	"testing"
	"time"
)

func Test_Mine(t *testing.T) {
	tcfg, err := Config()
	if err != nil {
		t.Fatalf("error configing for test %v", err)
	}
	tcfg.Set("Influx.Enabled", true)
	tcfg.Set("Influx.IP", "192.168.0.29")
	tcfg.Set("Influx.DB", "testMine")
	tcfg.Set("Influx.Port", 8086)
	tcfg.Set("influx.User", nil)
	tcfg.Set("Influx.Pw", nil)
	tcfg.Set("Influx.FlushSec", 1*time.Second)

	ms, err := NewMineSession(tcfg)
	if err != nil {
		t.Fatalf("failed to configure mining session")
	}
	ms.ca.resetEnabled = false
	err = ms.ca.GetStatus()
	if err != nil {
		t.Fatalf("%+v\n", err)
	}

	if err = ms.ca.ResetCards(false); err != nil {
		t.Fatalf("error Resetting cards %v\n", err)
	}

	ms.Met.enabled = true
	ms.Met.refresh = 10 * time.Second
	ms.Met.db = "tuningRun"
	if err = ms.Met.Config(tcfg); err != nil {
		t.Logf("failed to config metrics client")
	}

	go ms.Met.backGroundWriter()

	tests := []struct {
		name    string
		ms      *MineSession
		wantErr bool
	}{
		{"int 100", ms, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ms.Mine()
			if (err != nil) && !tt.wantErr {
				t.Errorf("%v", err)
			}

		})
	}

}

func TestCheckStak(t *testing.T) {
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

	ms, err := NewMineSession(tcfg)
	if err != nil {
		t.Fatalf("%v", err)
	}

	ms.ca.resetEnabled = false
	err = ms.ca.GetStatus()
	if err != nil {
		t.Fatalf("%+v\n", err)
	}

	if err = ms.ca.ResetCards(false); err != nil {
		t.Fatalf("error Resetting cards %v\n", err)
	}

	tests := []struct {
		name      string
		pid       int
		wantErr   bool
		errString string
	}{
		{"int 100 fail ", 0, true, "stak has stopped responding"},
		{"int 100", 33433, true, "process does not exist"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ms.m.Process = &os.Process{}
			ms.m.Process.Pid = tt.pid
			// fmt.Printf("%+v\n", ms.m.Process)
			if err := ms.m.CheckStakProcess(); err != nil {
				if (tt.wantErr == true) && (err.Error() != tt.errString) {
					t.Errorf("CheckStakProcess() error = %v, wantErr %v", err, tt.wantErr)
				}

			}
		})
	}
}

func Test_checkInternet(t *testing.T) {
	tests := []struct {
		name     string
		timeout  time.Duration
		maxFails int
		wantErr  bool
	}{
		{"google.co.uk", 3 * time.Second, 3, false},
		{"google.nnz.www", time.Second, 2, true},
		{"google.co.uk", 1 * time.Millisecond, 3, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkInternet(tt.name, tt.timeout, tt.maxFails); (err != nil) != tt.wantErr {
				t.Errorf("checkInternet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMineSession_CheckApi(t *testing.T) {
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
	tcfg.Set("Core.Stak.Ip", "192.168.0.4") // todo
	ms, err := NewMineSession(tcfg)
	if err != nil {
		t.Fatalf("%v", err)
	}

	ms.Api.Monitor(ms.Met)
	defer ms.Api.StopMonitor(ms.Met)
	tests := []struct {
		name      string
		pid       int
		wantErr   bool
		errString string
	}{
		{"int 100 fail ", 0, true, "stak has stopped responding"},
		{"int 100", 33433, true, "process does not exist"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ms.m.Process = &os.Process{}
			ms.m.Process.Pid = tt.pid
			// fmt.Printf("%+v\n", ms.m.Process)
			if err := ms.Api.CheckApi(4, time.Millisecond); err != nil {
				if (tt.wantErr == true) && (err.Error() != tt.errString) {
					t.Errorf("CheckApi) error = %v, wantErr %v", err, tt.wantErr)
				}

			}
		})
	}
}
