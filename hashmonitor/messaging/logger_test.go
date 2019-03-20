package messaging

import (
	"testing"
	"time"

	"github.com/spf13/viper"
)

//
// func Test_logger_fileWriter(t *testing.T) {
// 	type args struct {
// 		c *viper.Viper
// 	}
// 	tests := []struct {
// 		name string
// 		c    *viper.Viper
// 	}{
// 		{"no config", c},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			l := NewLogger()
// 			l.fileWriter(tt.c)
// 			toFile(tt.c)
// 		})
// 	}
// }

func TestLogger_Send(t *testing.T) {
	c := viper.New()
	c.SetDefault("Core.Log.Dir", "logs")
	c.SetDefault("Core.Log.File", "hashmonitor.log")
	c.SetDefault("Core.Log.Rotate", true)

	l := NewLogger(c)

	go l.dispatcher()
	time.Sleep(time.Second * 1)
	defer close(l.control)
	m := msg{
		text:  "test message",
		queue: "device",
		mtype: "error",
		// 	attachment: "this is a test \n\t\tattachment \nstring \nit contains many lines\nhow's it looking",
		priority: 1,
		silent:   false,
	}

	for x := 0; x <= 10; x++ {
		if err := l.Send(m); err != nil {
			t.Errorf("Failed to send message")
		}
	}

	time.Sleep(2 * time.Second)
}
