package hashmonitor

import (
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

func Test_gpuConfParse(t *testing.T) {
	file := strings.Join([]string{".testcode", "heavy.txt"}, pathSep)
	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("Can't find File %v, %v", file, err)
	}
	defer f.Close()

	tst := NewAmdConfig()

	if err = tst.gpuConfParse(f); err != nil {
		t.Errorf("cardAlgo.gpuConfParse() error = %v", err)
	}
	t.Logf("%v", tst)
}

func Test_amdConf_amdIntTemplate(t *testing.T) {
	file := strings.Join([]string{".testcode", "heavy.txt"}, pathSep)
	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("Can't find File %v, %v", file, err)
	}
	defer f.Close()

	tst := NewAmdConfig()
	if err = tst.gpuConfParse(f); err != nil {
		t.Errorf("AmdConf.gpuConfParse() error = %v", err)
	}

	if amd, err := tst.amdIntTemplate(1, root+".testcode"+pathSep); err != nil {
		t.Errorf("AmdConf.amdIntTemplate() error = %v,%v ", err, amd)
	}

}

func TestNewAmdConfig(t *testing.T) {
	var tests []struct {
		name string
		want AmdConf
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAmdConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAmdConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAmdConf_gpuConfParse(t *testing.T) {
	type args struct {
		r io.ReadCloser
	}
	var tests []struct {
		name    string
		mc      *AmdConf
		args    args
		wantErr bool
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mc.gpuConfParse(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("AmdConf.gpuConfParse() error = %v, match %v", err, tt.wantErr)
			}
		})
	}
}

func TestAmdConf_amdIntTemplate(t *testing.T) {
	type args struct {
		interleave int
		dir        string
	}
	var tests []struct {
		name    string
		mc      *AmdConf
		args    args
		wantStr string
		wantErr bool
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr, err := tt.mc.amdIntTemplate(tt.args.interleave, tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("AmdConf.amdIntTemplate() error = %v, match %v", err, tt.wantErr)
				return
			}
			if gotStr != tt.wantStr {
				t.Errorf("AmdConf.amdIntTemplate() = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}
