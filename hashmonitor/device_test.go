package hashmonitor

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"
)

func TestDevCon(t *testing.T) {
	config, err := Config()
	if err != nil {
		t.Errorf("error getting config %v", err)
	}
	results, err := winCmd(config.GetString("Core.Stak.Dir"), "powershell devcon.exe status =display")
	if err != nil {
		t.Errorf("%v", err)
	}
	fmt.Println(results)
}

func TestGetStatus(t *testing.T) {
	config, _ := Config()
	cards := NewCardData()

	err := cards.GetStatus(config)
	if err == nil {
		for k, v := range *cards {
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
				t.Errorf("winElevationCheck() error = %v, wantErr %v", err, tt.wantErr)
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
	cards := NewCardData()
	err := cards.GetStatus(rstEnabled)
	if err != nil {
		t.Errorf("%+v\n", err)
	}
	type args struct {
		c     *viper.Viper
		force bool
	}

	tests := []struct {
		name    string
		ca      *CardData
		args    args
		wantErr bool
	}{
		{"rstEnabled noForce", cards, args{rstEnabled, false}, false},
		{"rstDisabled noForce", cards, args{rstDisabled, false}, true},
		{"rstEnabled force", cards, args{rstEnabled, true}, false},
		{"rstDisabled force", cards, args{rstDisabled, true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ca.ResetCards(tt.args.c, tt.args.force); (err != nil) != tt.wantErr {
				t.Errorf("CardData.ResetCards() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
