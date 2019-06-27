package profitSwitch

import (
	"fmt"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestGetMineCryptoNightStats(t *testing.T) {
	coins, _, err := GetMineCryptoNightStats(2000, 0)
	if err != nil {
		t.Errorf("failed to get stats %v", err)
	}

	for _, v := range coins {
		fmt.Printf("%v\n", v)
	}

}

func Test_rewards_CustomCalcs(t *testing.T) {
	v := viper.New()
	v.SetDefault("Profit.Calc.Conceal", 1.8)
	v.SetDefault("Profit.Calc.v2", 0.95)
	v.SetDefault("Profit.Calc.v4", 0.95)
	v.SetDefault("Profit.Calc.Heavy", 0.6)
	v.SetDefault("Profit.Calc.HeavyX", 0.6)
	v.SetDefault("Profit.Calc.Haven", 0.6)
	v.SetDefault("Profit.Calc.Saber", 0.6)
	v.SetDefault("Profit.Calc.Litev1", 2.5)
	v.SetDefault("Profit.Calc.Fastv2", 1.8)
	cu := NewCustomCalcs(v)

	con := make(map[string]float64)
	rew, con, err := GetMineCryptoNightStats(2000, 100)
	if err != nil {
		t.Fatalf("failed getting stats %v", err)
	}

	tests := []struct {
		name     string
		r        *rewards
		div, res float64
		wantErr  bool
	}{
		{"", &rew, 1, 1, false},
		{"", &rewards{{"CCX", "cryptonight-conceal", time.Now(), time.Now(), .0001143641}}, 1, 1, false},
		{"", &rewards{{"MSR", "cryptonight-fast-v2", time.Now(), time.Now(), .0001143641}}, 1, 1, false},
		{"", &rewards{{"CCI", "cryptonight-concean", time.Now(), time.Now(), .0001143641}}, 1, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			con["CryptonightV2V4Factor"] = 0.95
			con["CryptonightFastV2Factor"] = 1.8
			con["CryptonightLiteFactor"] = 2.5
			con["CryptonightHeavyFactor"] = 0.6
			// r := tt.r
			// fmt.Println(*r)
			err := tt.r.CustomCalcs(cu, con)
			if (err != nil) != tt.wantErr {
				t.Fatalf("%v", err)
			}
			// fmt.Println(*r)

		})
	}
}

func Test_customCalc_custCalc(t *testing.T) {
	v := viper.New()
	v.SetDefault("Profit.Calc.Heavy", 0.6)
	v.SetDefault("Profit.Calc.HeavyX", 0.6)
	v.SetDefault("Profit.Calc.Haven", 0.6)
	v.SetDefault("Profit.Calc.Saber", 0.6)
	v.SetDefault("Profit.Calc.v2", 0.95)
	v.SetDefault("Profit.Calc.v4", 0.95)
	v.SetDefault("Profit.Calc.Conceal", 1.8)
	v.SetDefault("Profit.Calc.Fastv2", 1.8)
	v.SetDefault("Profit.Calc.Litev1", 2.5)

	cu := NewCustomCalcs(v)
	co := map[string]float64{
		"CryptonightV2V4Factor":   0.95,
		"CryptonightFastV2Factor": 1.8,
		"CryptonightHeavyFactor":  0.6,
		"CryptonightLiteFactor":   2.5}
	type args struct {
		algo string
		con  map[string]float64
	}
	tests := []struct {
		name              string
		va                customCalc
		args              args
		wantDivider       float64
		wantErr, diverror bool
	}{

		{"heavyx", cu, args{"cryptonight-heavy-x", co}, 0.6, false, false},
		{"heavy", cu, args{"cryptonight-heavy", co}, 0.6, false, false},
		{"haven", cu, args{"cryptonight-haven", co}, 0.6, false, false},
		{"saber", cu, args{"cryptonight-saber", co}, 0.6, false, false},
		{"v2", cu, args{"cryptonight-v2", co}, 0.95, false, false},
		{"v4", cu, args{"cryptonight-v4", co}, 0.95, false, false},
		{"cryptonight", cu, args{"cryptonight", co}, 1, false, false},
		{"conceal", cu, args{"cryptonight-conceal", co}, 1.8, false, false},
		{"fastv2", cu, args{"cryptonight-fast-v2", co}, 1.8, false, false},
		{"litev1", cu, args{"cryptonight-lite-v1", co}, 2.5, false, false},

		{"conceal fail", cu, args{"cryptonight-conceal", co}, 1.6, false, true},
		{"unknown", cu, args{"cryptonight-fast-v3", co}, 1.8, true, true},
		{"unknown2", cu, args{"cryptonight-lite-vx", co}, 2.5, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotDivider, err := tt.va.custCalc(tt.args.algo, tt.args.con)
			if (err != nil) != tt.wantErr {
				t.Errorf("customCalc.custCalc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (gotDivider != tt.wantDivider) != tt.diverror {
				t.Errorf("customCalc.custCalc() = %v, want %v", gotDivider, tt.wantDivider)
			}

		})
	}
}
