package hashmonitor

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/gogap/logrus_mate"
	_ "github.com/gogap/logrus_mate/hooks/expander"
	_ "github.com/gogap/logrus_mate/hooks/file"
	"github.com/sirupsen/logrus"
	// _ "github.com/gogap/logrus_mate/hooks/slack"
)

func TestDefaultConfig(t *testing.T) {
	fn := "defaultconfigtest.conf"
	err := defaultLoggerConfig(fn)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// cleanup
	err = os.Remove(fn)
	if err != nil {
		t.Fatalf("failed to delete %v %v", fn, err)
	}
}

// func Test_configLogger(t *testing.T) {
//
// 	err := ConfigLogger()
// 	if err != nil {
// 		t.Fatalf("failed to configure logging \n%v", err)
// 	}
// }

var invalidconf = `hashmonitor{
level = "error"
	formatter.name = "text"
	formatter.options{
		force-colors = false
		disable-colors = false
		disable-timestamp = false
		full-timestamp = false
		timestamp-format = "2006-01-02 15:04:05"
		disable-sorting = (false
		}

hooks{
	expander{}
	file{
		filename = "hashmonitor.log"
		daily = true
		rotate = true
		}
	}
}
`

func TestConfigLogger(t *testing.T) {
	testfolder, err := ioutil.TempDir(".testcode", "log")
	if err != nil {
		t.Fatalf("%v", err)
	}
	fmt.Printf("using dir %v\n", testfolder)
	defer cleanupDir(testfolder)

	type args struct {
		force    bool
		existing bool
		valid    bool
	}
	tests := []struct {
		name string
		args
		wantErr bool
	}{
		{"missing", args{false, false, false}, false},
		{"missing force", args{true, false, false}, false},

		{"force existing", args{true, true, true}, false},
		{"no force existing", args{false, true, true}, false},

		{"force invalid", args{true, true, false}, false},
		{"no force invalid VALUE", args{false, true, false}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := ioutil.TempFile(testfolder, "*.conf")
			if err != nil {
				t.Errorf("error creating temp file %v", err)
			}
			fn := file.Name()

			if tt.args.existing {
				if tt.valid {
					_, err = file.WriteString(conf)
					if err != nil {
						t.Errorf("failed writing config to test file")
					}
				} else {
					_, err = file.WriteString(invalidconf)
					if err != nil {
						t.Errorf("failed writing config to test file")
					}
				}

				file.Close()
			} else {
				file.Close()
				err = os.Remove(file.Name())
				if err != nil {
					t.Logf("removal error %v", err)
				}
			}

			if err := ConfigLogger(fn, tt.force); (err != nil) != tt.wantErr {
				t.Errorf("%v:  %v, wantErr %v", tt.name, err, tt.wantErr)
			}

		})
	}

}

func TestSlackMessage(t *testing.T) {
	// if err := ConfigLogger("logging.conf", true); err != nil {
	// 	t.Fatal("failed configuring logger")
	// }
	l := logrus.StandardLogger()
	slack := `slack {
	level = "info"
	formatter.name = "text"
	formatter.options{
		force-colors = true
		disable-colors = false
		disable-timestamp = true
		full-timestamp = false
		timestamp-format = "2006-01-02 15:04:05"
		disable-sorting = false
		}

hooks{
	slack {
        url      = "https://hooks.slack.com/services/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"
        levels   = ["debug", "error", "info", "warn"]
        channel  = ""
        emoji    = ":rag:"
        username = "logrus_mate"
        }
	}
}`

	mate, err := logrus_mate.NewLogrusMate(logrus_mate.ConfigString(slack))
	if err != nil {
		t.Fatalf("failed to configure logrus_mate %v", err)
	}
	if err = mate.Hijack(l, "slack"); err != nil {
		t.Fatalf("failed to hijack logrus %v", err)
	}

	tests := []struct {
		name    string
		wantErr bool
	}{
		{"info", false},
		{"warn", false},
		{"error", false},
		{"fatal", false},
		{"debug", false},
		{"panic", false},

		{"attachment", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Log("recovered from log.panic\n")
				}
			}()
			// 	host, _ := os.Hostname()
			switch tt.name {
			case "info":
				l.Infof(tt.name)
			case "warn":
				l.Warnf(tt.name)
			case "error":
				l.Errorf(tt.name)
			case "fatal":
				// can't test this one
				// l.Fatalf("%v %v",tt.name, time.Now().String())
				fmt.Println(tt.name)
			case "debug":
				l.Debugf(tt.name)
			case "panic":
				l.Panicf(tt.name)
			case "attachment":
				l.WithField("testkey", "testvalue")

			}
			// 	if err != nil {
			// 		t.Errorf("failed to config logger")
			// 	}
			// c.Hijack("log","")
			// 	// if err := ConfigLogger(fn, tt.force); (err != nil) != tt.wantErr {
			// 	// 	t.Errorf("%v:  %v, wantErr %v", tt.name, err, tt.wantErr)
			// 	// }
			// fmt.Printf("%v",log)
		})
	}
	t.Log(time.Now())

}

func cleanupDir(d string) {
	fmt.Printf("removing %v\n", d)
	err := os.RemoveAll(d)
	if err != nil {
		fmt.Printf("cleanup failed %v\n", err)
	}
}
