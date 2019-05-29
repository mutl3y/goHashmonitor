package hashmonitor

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// func Test_testPlan_New(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		match bool
// 	}{
// 		{"amd20", false},
// 		{"amd/20", true},
// 		{"Sp3cial", false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tp := NewTestPlan(tt.name)
// 			if err := tp.NewFile(); (err != nil) != tt.match {
// 				t.Errorf("testPlan.NewFile() error = %v, match %v", err, tt.match)
// 			}
// 		})
// 	}
//
// 	for _, v := range tests {
// 		f := fmt.Sprintf("%v/%v.toml", tpDir, v.name)
// 		err := os.Remove(f)
// 		if err != nil && !v.match {
// 			t.Logf("Cleanup failed %v", err)
// 		}
// 	}
// }
func Test_testPlan_Save(t *testing.T) {
	type fields struct {
		name         string
		minIntensity int
		maxIntensity int
		runTimeSecs  time.Duration
		startTime    time.Time
		endTime      time.Time
		results      []intensity
		file         string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "", fields: fields{name: "amd20", minIntensity: 20, maxIntensity: 40, runTimeSecs: 60}, wantErr: false},
		{name: "", fields: fields{name: "amd/20", minIntensity: 20, maxIntensity: 40, runTimeSecs: 60}, wantErr: true},
		{name: "", fields: fields{name: "sp3cial", minIntensity: 20, maxIntensity: 20, runTimeSecs: 6}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := NewTestPlan(tt.fields.name)
			tp.MinIntensity = tt.fields.minIntensity
			tp.MaxIntensity = tt.fields.maxIntensity
			tp.RunTime = tt.fields.runTimeSecs
			if err := tp.Save(); (err != nil) != tt.wantErr {
				t.Errorf("testPlan.Save() error = %v, match %v", err, tt.wantErr)
			} else {
				err = os.Remove(tp.File)
				if err != nil {
					t.Logf("Cleanup failed %v", err)
				}
			}
		})
	}
}

func Test_testPlan_Load(t *testing.T) {
	tp := NewTestPlan("amd20")

	tests := []struct {
		name  string
		force bool
	}{
		{name: "", force: true},
		{name: "", force: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tp.Load(tt.force); err != nil {
				t.Fatalf("testPlan.Load() error = %v", err)
			} else {
				err = os.Remove(tp.File)
				if err != nil {
					t.Logf("Cleanup failed %v", err)
				}
			}
		})
	}
}

func Test_testPlan_Results(t *testing.T) {
	type fields struct {
		name         string
		minIntensity int
		maxIntensity int
		runTime      string
		startTime    time.Time
		endTime      time.Time
		results      []intensity
		file         string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "", fields: fields{name: "amd20", minIntensity: 20, maxIntensity: 40, runTime: "5s"}, wantErr: false},
		{name: "", fields: fields{name: "amd/20", minIntensity: 20, maxIntensity: 40, runTime: "5s"}, wantErr: true},
		{name: "", fields: fields{name: "sp3cial", minIntensity: 20, maxIntensity: 20, runTime: "5s"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := NewTestPlan(tt.fields.name)
			tp.MinIntensity = tt.fields.minIntensity
			tp.MaxIntensity = tt.fields.maxIntensity
			ti, err := time.ParseDuration(tt.fields.runTime)
			if err != nil {
				t.Fatalf("testPlan.Results().parsetime\n %v", err)
			}
			tp.RunTime = ti
			if err = tp.Load(false); (err != nil) != tt.wantErr {
				t.Fatalf("testPlan.Results() error loading file %v\n", err)

			}
			if err = tp.Results(); err != nil {
				t.Errorf("testPlan.Results() issue getting results %v %+v\n", err, tt)
			}

			for i, v := range tp.ResultsSlice {
				fmt.Printf("index: %v, value: %v", i, v)
			}
			if _, err = os.Stat(tp.File); err == nil {
				if err = os.Remove(tp.File); err != nil {
					t.Logf("Cleanup failed %v", err)

				}
			}

		})
	}
}

// func TestNewIntensityRun(t *testing.T) {
// 	type args struct {
// 		conf *AmdConf
// 	}
// 	var tests []struct {
// 		name string
// 		args args
// 		want *intensity
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := NewIntensityRun(tt.args.conf); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("NewIntensityRun() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
//
// func TestNewTestPlan(t *testing.T) {
// 	type args struct {
// 		n string
// 	}
// 	var tests []struct {
// 		name string
// 		args args
// 		want *testPlan
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := NewTestPlan(tt.args.n); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("NewTestPlan() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
//
// func Test_testPlan_Run(t *testing.T) {
// 	type args struct {
// 		c *viper.Viper
// 	}
// 	var tests []struct {
// 		name    string
// 		t       *testPlan
// 		args    args
// 		wantErr bool
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.t.Run(tt.args.c); (err != nil) != tt.wantErr {
// 				t.Errorf("testPlan.Run() error = %v, match %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
//
// func Test_newIntensity(t *testing.T) {
// 	var tests []struct {
// 		name string
// 		want intensity
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := newIntensity(); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("newIntensity() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
