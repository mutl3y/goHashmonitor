package hashmonitor

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

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

func cleanup(f chan *os.File, stats bool) {
	for fn := range f {
		// set perms
		err := os.Chmod(fn.Name(), 0222)
		if err != nil {
			fmt.Printf("failed chmod  %v\n", err)
			return
		}

		// return file stats
		if stats {
			stat, err := os.Stat(fn.Name())
			if err != nil {
				fmt.Printf("failed reading stats %v\n", err)
				return
			}
			fmt.Printf(" %v \n %+v", fn.Name(), stat)
		}

		// attempt deletion
		dir := filepath.Dir(fn.Name())
		err = os.RemoveAll(dir)
		if err == nil {
			return
		}
		fmt.Printf("failed removing %v %v\n", fn.Name(), err) // handle errors

		b, err := winCmd("", "powershell rm "+fn.Name())
		if err != nil {
			fmt.Printf("error removing %v %v\n", fn.Name(), err)
			return
		}
		fmt.Printf("%s", b)
	}
}

func TestConfig(t *testing.T) {

	type pair struct {
		name   string
		value  interface{}
		result interface{}
	}

	tests := []struct {
		name string
		pair
		wantErr bool
	}{
		{"int", pair{"test", 666, 666}, false},
		// {"decimal", pair{"test", 21.1, 21.1}, false},
		// {"negative int", pair{"test", -33, -33}, false},
		// {"string", pair{"test", "special", "special"}, false},
		// {"bool", pair{"test", true, true}, false},
		// {"time", pair{"test", time.Second * 3, time.Second * 3}, false},
		// {"wrong_int", pair{"test", 666, 667}, true},
		// {"wrong_decimal", pair{"test", 21.1, "21.1"}, true},
		// {"wrong_negative int", pair{"test", -33, "-33"}, true},
		// {"wrong_string", pair{"test", "special", 666}, true},
		// {"wrong_bool", pair{"test", true, false}, true},
		// {"wrong_time", pair{"test", time.Second * 3, time.Second * 4}, true},
	}

	files := make(chan *os.File, 100)
	go cleanup(files, false)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := ioutil.TempFile(".testcode"+pathSep+"configs", "*.yaml")
			if err != nil {
				fmt.Printf("error creating temp file %v", err)
			}
			file.Close()
			n := file.Name()
			// new config
			v := viper.New()
			v.SetConfigFile(n)

			// set the value
			v.Set(tt.pair.name, tt.pair.value)

			// write to disk
			err = v.WriteConfigAs(v.ConfigFileUsed())
			if err != nil {
				t.Errorf("write config failed %v", err)
			}

			// new config
			v2 := viper.New()
			v2.SetConfigFile(n)

			// read in from disk
			err = v2.ReadInConfig()
			if err != nil {
				t.Errorf("failed reading config %v", err) // handle errors
			}

			// get the value and compare
			res := v2.Get(tt.pair.name)
			if (res != tt.pair.result) != tt.wantErr {
				t.Errorf("%v want %v %T got %v %T , wantErr %v", tt.name, res, res, tt.pair.value, tt.pair.value, tt.wantErr)

			}

			files <- file
		})
	}

	close(files)
	time.Sleep(3 * time.Second)
}
