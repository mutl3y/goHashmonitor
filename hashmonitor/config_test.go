package hashmonitor

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
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
			fmt.Printf("returning %v", err)
			return
		}

		var inUse = errors.New("The process cannot access the file because it is being used by another process.")
		if err != inUse {
			// 		fmt.Printf("file in use %v \nWill cleanup next run\n", fn.Name()) // handle errors

		} else {
			fmt.Printf("error during cleanup %v", err)
		}

		// b, err := winCmd("", "powershell rm "+fn.Name())
		// if err != nil {
		// 	fmt.Printf("error removing %v %v\n", fn.Name(), err)
		// 	return
		// }
		// fmt.Printf("%s", b)
	}
}

func TestConfig(t *testing.T) {
	dir := ".testcode" + pathSep + "configTests"
	err := os.MkdirAll(dir, 644)
	if err != nil {
		t.Fatal("failed creating test dir")
	}
	type pair struct {
		value  interface{}
		result interface{}
	}

	tests := []struct {
		name string
		pair
		wantErr bool
	}{
		{"int", pair{666, 666}, false},
		{"decimal", pair{21.1, 21.1}, false},
		{"negative int", pair{-33, -33}, false},
		{"string", pair{"special", "special"}, false},
		{"bool", pair{true, true}, false},
		{"time", pair{time.Minute * 3, time.Minute * 3}, false},
		{"wrong_int", pair{666, 667}, true},
		{"wrong_decimal", pair{21.1, "21.1"}, true},
		{"wrong_negative int", pair{-33, "-33"}, true},
		{"wrong_string", pair{"special", 666}, true},
		{"wrong_bool", pair{true, false}, true},
		{"wrong_time", pair{time.Second * 3, time.Second * 4}, true},
	}

	files := make(chan *os.File, 100)

	go cleanup(files, false)
	defer close(files)
	wg := sync.WaitGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wg.Add(1)
			file, err := ioutil.TempFile(dir, "*.yaml")
			if err != nil {
				fmt.Printf("error creating temp file %v", err)
			}
			file.Close()
			n := file.Name()
			// new config
			v := viper.New()
			v.SetConfigFile(n)

			// set the value
			v.Set(tt.name, tt.pair.value)

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

			var res interface{}
			// get the value and compare
			if strings.Contains(tt.name, "time") {
				res = v2.GetDuration(tt.name)
			} else {
				res = v2.Get(tt.name)

			}
			if (res != tt.pair.result) != tt.wantErr {
				t.Logf("%v want %v %T got %v %T , wantErr %v", tt.name, tt.pair.value, tt.pair.value, res, res, tt.wantErr)
				// 		t.Logf("%v %T", res2, res2)
			}
			files <- file
			wg.Done()
		})
	}
	wg.Wait()

}
