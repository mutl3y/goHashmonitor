package hashmonitor

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	_ "github.com/gogap/logrus_mate/hooks/expander"
	_ "github.com/gogap/logrus_mate/hooks/file"
	_ "github.com/gogap/logrus_mate/hooks/slack"
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

func TestConfigLogger(t *testing.T) {
	testfolder, err := ioutil.TempDir(".testcode", "")
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

		{"no force existing", args{false, true, true}, false},
		{"force existing", args{true, true, true}, false},

		{"force existing invalid", args{true, true, false}, false},
		{"no force invalid", args{false, true, false}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := ioutil.TempFile(testfolder, "*.conf")
			if err != nil {
				t.Logf("error creating temp file %v", err)
			}
			fn := file.Name()

			if tt.args.existing {
				if tt.valid {
					_, err = file.WriteString(conf)
					if err != nil {
						t.Errorf("failed writing config to test file")
					}
				}
				defer file.Close()
			} else {
				file.Close()
				err = os.Remove(file.Name())
				if err != nil {
					t.Logf("removal error %v", err)
				}
			}

			if err := ConfigLogger(fn, tt.force); (err != nil) != tt.wantErr {
				t.Logf("%v:  %v, wantErr %v", tt.name, err, tt.wantErr)
			}

		})
	}

}

func cleanupDir(d string) {
	fmt.Printf("removing %v\n", d)
	err := os.RemoveAll(d)
	if err != nil {
		fmt.Printf("cleanup failed %v\n", err)
	}
}
