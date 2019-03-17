package hashmonitor

import (
	"testing"
)

func TestSendSlackMessage(t *testing.T) {
	config, _ := Config()
	cl, _ := NewSlackConfig(config)
	type args struct {
		msg    string
		slType string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"error message", args{"test error message", "error"}, false},
		{"warn message", args{"test warn message", "warn"}, false},
		{"info message", args{"test info message", "info"}, false},
		{"default message", args{"test default message", "other"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cl.SendMessage(tt.args.msg, tt.args.slType); (err != nil) != tt.wantErr {
				t.Errorf("SendSlackMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
