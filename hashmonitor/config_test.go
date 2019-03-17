package hashmonitor

import (
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

func Test_defaultConfig(t *testing.T) {
	tests := []struct {
		name    string
		want    *viper.Viper
		wantErr bool
	}{
		{"Test type", viper.New(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			if cfg == nil {
				t.Errorf("DefaultConfig() failed")
				return
			}

		})
	}
}

func TestConfig(t *testing.T) {
	var tests []struct {
		name    string
		want    *viper.Viper
		wantErr bool
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Config()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Config() = %v, want %v", got, tt.want)
			}
		})
	}
}
