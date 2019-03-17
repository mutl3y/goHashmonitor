package hashmonitor

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	tpDir = "./testPlans"
)

type testPlan struct {
	Name, File                   string
	MinIntensity, MaxIntensity   int
	RunTime                      time.Duration
	ResultsSlice                 []intensity
	Creation, StartTime, EndTime time.Time
}

type TestPlan interface {
	Load() error
	Save() error
	Run() error
	Results() error
}

func NewTestPlan(n string) *testPlan {
	tp := new(testPlan)
	tp.Name = n
	tp.File = fmt.Sprintf("%v/%v.toml", tpDir, n)
	tp.ResultsSlice = make([]intensity, 0, 10)
	tp.Creation = time.Now()
	return tp
}

// Load Loads a testplan from a file, pass true to force new
func (t *testPlan) Load(force bool) (err error) {
	if _, err = os.Stat(tpDir); os.IsNotExist(err) {
		err := os.Mkdir(tpDir, 666)
		if err != nil {
			log.Fatalf("failed to mkdir %v", tpDir)
		}
	}

	if _, err = os.Stat(t.File); os.IsNotExist(err) || force {
		err = t.Save()
		if err != nil {
			return fmt.Errorf("failed saving new file %s error was %v", t.File, err)
		}
	}
	fp, err := os.OpenFile(t.File, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s error was %v", t.File, err)

	}

	if _, err = toml.DecodeFile(t.File, &t); err != nil {
		return fmt.Errorf("failed to decode file %s error was %v", t.File, err)
	}
	err = fp.Close()
	if err != nil {
		log.Fatalf("failed to close file %v", fp.Name())
	}
	return err
}

// Save testPlan to file with overwrite
// will be used during runs to save progress for restart and diagnosis
func (t *testPlan) Save() (err error) {
	if _, err = os.Stat(tpDir); os.IsNotExist(err) {
		err = os.Mkdir(tpDir, 666)

		if err != nil {
			log.Fatalf("failed to close file")
		}
	}
	fp, err := os.OpenFile(t.File, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file for writting %v, error was %v", t.File, err)
	}

	b := new(bytes.Buffer)
	e := toml.NewEncoder(b)
	err = e.Encode(&t)
	if err != nil {
		return fmt.Errorf("failed to testplan %v", err)
	}

	_, err = fp.Write(b.Bytes())
	err = fp.Close()
	if err != nil {
		log.Fatalf("failed to close file %v", fp.Name())
	}
	return
}

func (t *testPlan) Run(c *viper.Viper) (err error) {

	stakdir := c.GetString("Core.Stak.Dir") + string(os.PathSeparator)
	amdText := stakdir + "amd.txt"
	f, err := os.OpenFile(amdText, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("tp.Run.open %v", err)
	}
	//noinspection ALL
	defer f.Close()

	tst := NewAmdConfig()
	if err = tst.gpuConfParse(f); err != nil {
		return fmt.Errorf("tp.run.conf %v", err)
	}

	var amdFile string
	if amdFile, err = tst.amdIntTemplate(20, stakdir); err != nil {
		return fmt.Errorf("tp.run.template %v", err)
	}

	m := NewMiner()
	ctx, err := m.ConfigMiner(c)
	if err != nil {
		return fmt.Errorf("tp.run.config: %v", err)
	}

	m.config.amdFile = amdFile
	ti := int64(t.RunTime / time.Second)
	fmt.Printf("*****************************************************************%v", ti)
	m.config.args = append(m.config.args, "--benchmark", "cryptonight_heavy", "--benchwait", "5", "--benchwork", "10", "--amd", amdText)
	fmt.Printf("%v", m.config)
	err = m.StartMining(ctx)
	if err != nil {
		return fmt.Errorf("intensity.run.start: %v", err)
	}
	time.AfterFunc(60*time.Second, func() {
		err := m.StopMining()
		if err != nil {

		}
	})
	_, err = m.Process.Wait()
	if err != nil {
		log.Fatalf("failed to wait for process %v", m.Process.Pid)
	}

	return
}

// Results Displays raw results to screen
// todo expand
func (t *testPlan) Results() (err error) {
	fmt.Printf("results: %+v\n", t.ResultsSlice)
	return nil
}

type TestResults struct {
	Date      time.Time
	Intensity int
	Hashrate  float64
	MsgPerMin float64
}

type intensity struct {
	Results []TestResults
	Config  AmdConf
}

func newIntensity() intensity {
	i := intensity{}
	i.Results = make([]TestResults, 0, 100)
	return i
}

type Intensity interface {
	LoadResults() error
	RunTest() error
}

func NewIntensityRun(conf *AmdConf) (i *intensity) {
	i = &intensity{}
	i.Results = make([]TestResults, 0, 60)
	return
}
