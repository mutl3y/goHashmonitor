package hashmonitor

import (
	"bufio"
	"context"
	"fmt"
	"github.com/SkyrisBactera/pkill"
	"io"
	"os"
	"os/exec"
	"runtime"
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
	Up, Running bool

	Stop       context.CancelFunc
	Process    *os.Process
	StdOutPipe *io.ReadCloser
	ctx        context.Context
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

func OSSettings() {

	switch Os := runtime.GOOS; {
	case Os == "windows":

	case Os == "linux":
		cmd := exec.Command("bash", "-c", "sudo sysctl -w vm.nr_hugepages=128")
		err := cmd.Run()
		if err != nil {
			fmt.Println("Failed to setup hugepages", err)
		}

	default:
		log.Fatalf("Config() OS not supported")
	}
}

// ConfigMiner creates the base config, attaches a context and embeds a cancel func in the struct
func (ms *miner) ConfigMiner(c *viper.Viper) error {
	ms.config.dir = c.GetString("Core.Stak.Dir")
	if ms.config.dir == "" {
		return fmt.Errorf("stak Directory Not Specified")
	}
	ms.config.exe = c.GetString("Core.Stak.Exe")
	if ms.config.exe == "" {
		return fmt.Errorf("stak Executable Not Specified")
	}
	ms.config.args = c.GetStringSlice("Core.Stak.Args")
	ms.config.startAttempts = c.GetInt("Core.Stak.Start_Attempts")
	ms.tools = c.GetStringSlice("Core.Stak.Tools")
	ms.ctx, ms.Stop = context.WithCancel(context.Background())
	ms.Running = true
	return nil
}

// StartMining a configured miner
func (ms *miner) StartMining() error {
	OSSettings()

	cmd := exec.CommandContext(ms.ctx, ms.config.exe, ms.config.args...)
	cmd.Dir = ms.config.dir
	stdPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdOut pipe, %v", err)
	}

	if Debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err = cmd.Start(); err != nil {
		debug("%+v", cmd)
		return fmt.Errorf("failed to start mining process, %v", err)
	}

	ms.Process = cmd.Process
	ms.StdOutPipe = &stdPipe
	debug("Starting STAK process ID %v", ms.Process.Pid)
	fmt.Println("taking a breath, allowing stak to start, 10 seconds ...")

	time.Sleep(10 * time.Second)
	return err
}
func (ms *miner) StopMining() error {
	debug("killing process id: %v", ms.Process.Pid)
	ms.Stop()
	if ms.Up {
		ErrAccess := fmt.Errorf("access is denied")
		if err := ms.Process.Kill(); err != nil && err != ErrAccess {
			debug("failed to kill miner %v", err)
			return err
		}
	}
	exe := ms.config.exe
	if exe == "" {
		return fmt.Errorf("exe not specified")
	}
	exe = strings.ReplaceAll(exe, "./", "")

	_, err := pkill.Pkill(exe)
	if err != nil && (err.Error() != "exit status 1") {
		debug("pkill error %v %v", exe, err)
	}
	ms.Running = false

	return nil
}

func (ms *miner) killStak() error {
	exe := ms.config.exe
	if exe == "" {
		return fmt.Errorf("exe not specified")
	}
	exe = strings.ReplaceAll(exe, "./", "")

	_, _ = pkill.Pkill(exe)

	return nil
}

func (ms *miner) ConsoleMetrics(met *metrics) {

	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf("failed to set hostname %v", err)
	}

	scanner := bufio.NewScanner(*ms.StdOutPipe)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m, err := conParse(scanner.Bytes())
		if err != nil && err.Error() != "no match" {
			log.Errorf("Error parsing %v\n", err)

		}
		if len(m) > 0 {
			tags := map[string]string{"server": hostname}
			if m["TAG"] != nil {
				tags["type"] = fmt.Sprintf("%v", m["TAG"])
				delete(m, "TAG")
			}
			if m["gpu"] != nil {
				tags["gpu"] = fmt.Sprintf("%v", m["gpu"])
				delete(m, "gpu")
			}
			if m["thread"] != nil {
				tags["thread"] = fmt.Sprintf("%v", m["thread"])
				delete(m, "thread")
			}
			if err := met.Write("consoleMetrics", tags, m); err != nil {
				debug("console metrics error %v", err)
			}
			debug("ConsoleMetrics %+v %+v", tags, m)
		}
		if !ms.Running {
			fmt.Println("miner not running")
			break
		}
	}
	if err = scanner.Err(); err != nil {
		log.Errorf("Invalid input: %s", err)

	}

	return
}

func conParse(b []byte) (m map[string]interface{}, err error) {
	m = map[string]interface{}{}
	s := string(b)
	// fmt.Println(s)
	if !strings.HasPrefix(s, "[") {
		return m, fmt.Errorf("no match")
	}

	// grab datestamp
	// d := s[1:20]

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
			log.Errorf("conParse error length? %v", s)
		}

		switch s[:12] {
		case "OpenCL Inter":
			md, err := interleaveFilter(s[18:])
			if err != nil {
				log.Errorf("conParse interleave decoding error %v\n", err)

			}
			for k, v := range md {
				m[k] = v
			}

		case "OpenCL devic":
			debug("OpenCL device %v", s)

		default:
			if strings.Contains(s, "auto-tune validate") {
				md, err := autotuneFilter(s)
				if err != nil {
					log.Errorf("conParse autotune decoding error %v\n", err)

				}
				for k, v := range md {
					m[k] = v
				}
			} else if strings.Contains(s, "lock intensity at") {
				md, err := lockFilter(s)
				if err != nil {
					log.Errorf("conParse lock decoding error %v\n", err)

				}
				for k, v := range md {
					m[k] = v
				}
			} else {

				debug("conParse unparsed openCl %v\n", s)
			}
		}
	case "Mini":
		debug("algorithm \t%v", s[13:])

	// discarded and non printed below
	case "Devi": // fmt.Printf("device \t\t%v\n", s)
	case "WARN": // fmt.Printf("warning \t%v\n", s)
	case "Fast": // fmt.Printf("connecting\t%v\n", s)
	case "Pool": // fmt.Printf("pool \t\t%v\n", s)
	case "Diff": // fmt.Printf("difficulty \t%v\n", s)
	case "Comp": // fmt.Printf("compiling\t%v\n", s)
	case "Star": // fmt.Printf("startup \t%v\n", s)
	case "New ": // fmt.Printf("new block %v\n", s)
	case "Swit": // fmt.Printf("switch too %v\n", s)
	case "hwlo": // fmt.Printf("hwloc %v\n", s)
	case "Resu":
		debug("result %v\n", s)
	default:
		debug("Unparsed Message %v\n", s)
	}
	return m, nil
}
func parseStakUnixTime(s string) (int64, error) {
	date, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return 0, err
	}
	return date.Unix(), err

}

func interleaveFilter(s string) (m map[string]interface{}, err error) {
	// `<gpu id>|<thread id on the gpu>: <last delay>/<average calculation per hash bunch> ms - <interleave value>`
	m = make(map[string]interface{})

	fields := strings.Fields(s)
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		err = fmt.Errorf("dsddsdfsd")
	// 	}
	// }()

	gpuThread := strings.Split(fields[0][:len(fields[0])-1], "|")
	if len(gpuThread) > 0 {
		if m["gpu"], err = strconv.ParseInt(gpuThread[0], 0, 64); err != nil {
			return m, fmt.Errorf("gpu %v", err)
		}

		if m["thread"], err = strconv.ParseInt(gpuThread[1], 0, 64); err != nil {
			return m, fmt.Errorf("thread %v", err)
		}
	}

	// extract last and average
	laPair := strings.Split(fields[1], "/")
	if m["last"], err = strconv.ParseInt(laPair[0], 0, 64); err != nil {
		return m, fmt.Errorf("last %v", err)
	}
	if m["average"], err = strconv.ParseFloat(laPair[1], 64); err != nil {
		return m, fmt.Errorf("average %v", err)
	}

	if len(fields) >= 5 {
		if m["interleave"], err = strconv.ParseFloat(fields[4], 64); err != nil {
			return m, fmt.Errorf("interleave %v", err)
		}
	}
	m["TAG"] = "interleave_event"
	debug("interleave filter %v", fields)
	return m, nil

}

type autoTuneStat struct {
	gpu, thread, intensity int64
}

func autotuneFilter(s string) (m map[string]interface{}, err error) {
	// `<gpu id>|<thread id on the gpu>: <last delay>/<average calculation per hash bunch> ms - <interleave value>`
	m = make(map[string]interface{})

	fields := strings.Fields(s)
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		err = fmt.Errorf("dsddsdfsd")
	// 	}
	// }()

	gpuThread := strings.Split(fields[1][:len(fields[1])-1], "|")
	gpu, err := strconv.ParseInt(gpuThread[0], 0, 64)
	if err != nil {
		return m, fmt.Errorf("gpu %v", err)
	}
	m["gpu"] = gpu

	thread, err := strconv.ParseInt(gpuThread[1], 0, 64)
	if err != nil {
		return m, fmt.Errorf("thread %v", err)
	}
	m["thread"] = thread
	// extract last and average
	intPair := strings.Split(fields[5], "|")
	intensity, err := strconv.ParseInt(intPair[0], 0, 64)
	if err != nil {
		return m, fmt.Errorf("new %v", err)
	}
	m["newIntensity"] = intensity
	if m["oldIntensity"], err = strconv.ParseInt(intPair[1], 0, 64); err != nil {
		return m, fmt.Errorf("old %v", err)
	}
	m["TAG"] = "autotune_event"

	LockCounter.mu.Lock()
	id := fmt.Sprintf("ID%v%v", gpu, thread)
	LockCounter.threads[id] = autoTuneStat{gpu, thread, intensity}
	LockCounter.mu.Unlock()

	return m, nil

}

func lockFilter(s string) (m map[string]interface{}, err error) {
	// `<gpu id>|<thread id on the gpu>: <last delay>/<average calculation per hash bunch> ms - <interleave value>`
	m = make(map[string]interface{})

	fields := strings.Fields(s)
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		err = fmt.Errorf("dsddsdfsd")
	// 	}
	// }()

	gpuThread := strings.Split(fields[1][:len(fields[1])-1], "|")
	if m["gpu"], err = strconv.ParseInt(gpuThread[0], 0, 64); err != nil {
		return m, fmt.Errorf("gpu %v", err)
	}

	if m["thread"], err = strconv.ParseInt(gpuThread[1], 0, 64); err != nil {
		return m, fmt.Errorf("thread %v", err)
	}

	if m["lockIntensity"], err = strconv.ParseInt(fields[5], 0, 64); err != nil {
		return m, fmt.Errorf("new %v", err)
	}

	m["TAG"] = "intensityLock_event"
	LockCounter.mu.Lock()
	LockCounter.Counter += 1
	LockCounter.mu.Unlock()
	return m, nil

}

// ResetCards()
// func (ms *miner) ResetCards() {}
//
// func (ms *miner) TuneIntensity() {
//
// }
