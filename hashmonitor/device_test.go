package hashmonitor

import (
	"fmt"
	"testing"
)

func TestDevCon(t *testing.T) {
	config, err := Config()
	if err != nil {
		t.Errorf("error getting config %v", err)
	}
	res, err := winCmd(config.GetString("Core.Stak.Dir"), "powershell devcon.exe status =display")
	if err != nil {
		t.Errorf("%v", err)
	}
	fmt.Println(res)
}

func TestGetStatus(t *testing.T) {
	config, _ := Config()
	cd := NewCardData(config)

	err := cd.GetStatus()
	if err == nil {
		for k, v := range cd.cards {
			fmt.Printf("Card-%v %+v %+v\n", k, v.name, v.running)
		}
	}
}

func Test_winElevationCheck(t *testing.T) {
	tests := []struct {
		name    string
		want    bool
		wantErr bool
	}{
		{"", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := winElevationCheck()
			if (err != nil) != tt.wantErr {
				t.Errorf("winElevationCheck() error = %v, match %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("winElevationCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCardData_ResetCards(t *testing.T) {
	rstEnabled, _ := Config()
	rstDisabled, _ := Config()
	rstEnabled.Set("Device.Reset.Enabled", true)
	rstDisabled.Set("Device.Reset.Enabled", false)
	rstEnabled.Set("Core.Stak.Dir", "xmr-stak")
	rstDisabled.Set("Core.Stak.Dir", "xmr-stak")

	enabled := NewCardData(rstEnabled)
	disabled := NewCardData(rstDisabled)
	// err := ConfigLogger("logging.amdConf",false)

	err := enabled.GetStatus()
	if err != nil {
		t.Errorf("%+v\n", err)
	}
	err = disabled.GetStatus()
	if err != nil {
		t.Errorf("%+v\n", err)
	}

	tests := []struct {
		name    string
		ca      CardData
		force   bool
		wantErr bool
	}{
		{"rstEnabled noForce", *enabled, false, false},
		{"rstDisabled noForce", *disabled, false, false},
		{"rstEnabled force", *enabled, true, false},
		{"rstDisabled force", *disabled, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := tt.ca
			if err = ca.ResetCards(tt.force); (err != nil) != tt.wantErr {
				fmt.Printf("\n test config \n %+v\n", ca)
				t.Errorf("CardData.ResetCards() error = %v, match %v", err, tt.wantErr)
			}
		})
	}
}
