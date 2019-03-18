package messaging

import (
	"testing"

	"github.com/spf13/viper"
)

func Test_toFile(t *testing.T) {
	c := viper.New()
	c.SetDefault("Core.Log.Dir", "logs")
	c.SetDefault("Core.Log.File", "hashmonitor.log")
	c.SetDefault("Core.Log.Rotate", true)
	type args struct {
		c *viper.Viper
	}
	tests := []struct {
		name string
		c    *viper.Viper
	}{
		{"no config", c},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toFile(tt.c)
		})
	}
}
