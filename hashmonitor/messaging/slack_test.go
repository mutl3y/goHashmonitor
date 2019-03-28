package messaging

import (
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestSendSlackMessage(t *testing.T) {
	cfg := viper.New()
	cfg.SetDefault("Slack.Url", "https://hooks.slack.com/services/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")
	cfg.SetDefault("Slack.Username", "unsetUserName")
	cfg.SetDefault("Slack.PeriodicReporting.Enabled", false)
	cfg.SetDefault("Slack.PeriodicReporting.Minutes", 10)
	cfg.SetDefault("Slack.MessageWindow", time.Duration(time.Second*30))
	cfg.SetDefault("Slack.Verbosity", 2)
	cl, _ := NewSlackConfig(cfg)
	type args struct {
		msg    string
		slType string
		ts     int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"error message", args{"test error message", "error", 0}, false},
		{"warn message", args{"test warn message", "warn", 0}, false},
		{"info message", args{"test info message", "info", 0}, false},
		{"default message", args{"test default message", "other", int(time.Now().Unix())}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cl.SendMessage(tt.args.msg, tt.args.slType, tt.args.ts); (err != nil) != tt.wantErr {
				t.Errorf("SendSlackMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
