package hashmonitor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// var algos = []string{
// 	"aeon7",
// 	"bbscoin",
// 	"bittube",
// 	"cryptonight",
// 	"cryptonight_bittube2",
// 	"cryptonight_masari",
// 	"cryptonight_haven",
// 	"cryptonight_heavy",
// 	"cryptonight_lite",
// 	"cryptonight_lite_v7",
// 	"cryptonight_lite_v7_xor",
// 	"cryptonight_superfast",
// 	"cryptonight_v7",
// 	"cryptonight_v8",
// 	"cryptonight_v7_stellite",
// 	"freehaven",
// 	"graft",
// 	"haven",
// 	"intense",
// 	"masari",
// 	"monero",
// 	"qrl",
// 	"ryo",
// 	"stellite",
// 	"turtlecoin",
// }

type miner struct {
	config struct {
		args          []string
		amdFile       string
		dir           string
		exe           string
		startAttempts int
	}
	pool struct {
		url, tlsurl, user string
		rigid, pass, algo string
		nicehash          bool
	}
	tools []string
	// 	signal     chan struct{}
	Up         bool
	Stop       context.CancelFunc
	Process    *os.Process
	StdOutPipe *io.ReadCloser
}

type Miner interface {
	ConfigMiner(cfg *viper.Viper) error
	StartMining() error
	StopMining() error
}

func NewMiner() *miner {
	m := new(miner)
	return m
}

// ConfigMiner creates the base config and attaches a context and embeds a cancel func
func (ms *miner) ConfigMiner(c *viper.Viper) (context.Context, error) {
	ms.config.dir = c.GetString("Core.Stak.Dir")
	if ms.config.dir == "" {
		return nil, fmt.Errorf("stak Directory Not Specified")
	}
	ms.config.exe = c.GetString("Core.Stak.Exe")
	if ms.config.exe == "" {
		return nil, fmt.Errorf("stak Executable Not Specified")
	}
	ms.config.args = c.GetStringSlice("Core.Stak.Args")
	ms.config.startAttempts = c.GetInt("Core.Stak.Start_Attempts")
	ms.tools = c.GetStringSlice("Core.Stak.Tools")
	ctx, Stop := context.WithCancel(context.Background())
	ms.Stop = Stop
	return ctx, nil
}

// StartMining a configured miner
func (ms *miner) StartMining(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, ms.config.exe, ms.config.args...)
	cmd.Dir = ms.config.dir
	stdPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdOut pipe, %v", err)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Printf("%+v", cmd)
		return fmt.Errorf("failed to start mining process, %v", err)
	}

	ms.Process = cmd.Process
	ms.StdOutPipe = &stdPipe
	return err
}

func (ms *miner) ConsoleMetrics() error {
	scanner := bufio.NewScanner(*ms.StdOutPipe)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		err := conParse(scanner.Bytes())
		if err != nil && err.Error() != "no match" {
			fmt.Printf("Error parsing %v\n", err)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Invalid input: %s", err)
	}
	return nil
}

func conParse(b []byte) error {
	s := string(b)
	if !strings.HasPrefix(s, "[") {
		return fmt.Errorf("no match")
	}

	// grab datestamp
	d := s[1:20]

	// grab message
	s = s[24:]

	// grab first 4 characters for operation matching
	switch s[:4] {
	case "Open":
		// truncate
		if len(s) >= 50 {
			s = s[:50]
		}
		if len(s) <= 12 {
			log.Fatalf("conParse error length? %v", s)
		}

		switch s[:12] {
		case "OpenCL Inter":
			unixTime, err := parseStakUnixTime(d)
			if err != nil {
				unixTime = 0
			}
			if err := interleaveFilter(s[18:], unixTime); err != nil {
				fmt.Printf("conParse decoding error %v\n", err)
			}
		case "OpenCL devic":
			// fmt.Printf("device \Date\Date%v\n", s)
		default:
			fmt.Printf("conParse unparsed openCl %v\n", s)

		}
	case "Mini":
		fmt.Printf("algorithm \t%v\n", s[13:])

	// discarded and non printed below
	case "Devi": // fmt.Printf("device \t\t%v\n", s)
	case "WARN": // fmt.Printf("warning \t%v\n", s)
	case "Fast": // fmt.Printf("connecting\t%v\n", s)
	case "Pool": // fmt.Printf("pool \t\t%v\n", s)
	case "Diff": // fmt.Printf("difficulty \t%v\n", s)
	case "Comp": // fmt.Printf("compiling\t%v\n", s)
	case "Star": // fmt.Printf("startup \t%v\n", s)
	case "New ": // fmt.Printf("new block %v\n", s)
	case "Resu":
		fmt.Printf("result %v\n", s)
	default:
		fmt.Printf("Unparsed Message %v\n", s)
	}

	return nil
}
func parseStakUnixTime(s string) (int64, error) {
	date, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return 0, err
	}
	return date.Unix(), err

}

func interleaveFilter(s string, d int64) (err error) {
	// `<gpu id>|<thread id on the gpu>: <last delay>/<average calculation per hash bunch> ms - <interleave value>`
	type it struct {
		date                int64
		gpu, thread, last   int64
		average, interleave float64
	}

	m := it{}

	fields := strings.Fields(s)

	m.date = d

	gpuThread := strings.Split(fields[0][:len(fields[0])-1], "|")
	m.gpu, err = strconv.ParseInt(gpuThread[0], 0, 64)
	if err != nil {
		return fmt.Errorf("gpu %v", err)
	}
	m.thread, err = strconv.ParseInt(gpuThread[1], 0, 64)
	if err != nil {
		return fmt.Errorf("thread %v", err)
	}

	// extract last and average
	laPair := strings.Split(fields[1], "/")
	m.last, err = strconv.ParseInt(laPair[0], 0, 64)
	if err != nil {
		return fmt.Errorf("last %v", err)
	}
	m.average, err = strconv.ParseFloat(laPair[1], 64)
	if err != nil {
		return fmt.Errorf("average %v", err)
	}

	if len(fields) >= 5 {
		m.interleave, err = strconv.ParseFloat(fields[4], 64)
		if err != nil {
			return fmt.Errorf("interleave %v", err)
		}
	}

	fmt.Printf("%+v\n", m)
	return nil

}

func (ms *miner) StopMining() error {
	ms.Stop()
	fmt.Printf("killing process id: %v", ms.Process.Pid)
	err := ms.Process.Kill()

	return err
}

// ResetCards()
// func (ms *miner) ResetCards() {}
//
// func (ms *miner) TuneIntensity() {
//
// }
