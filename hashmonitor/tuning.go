package hashmonitor

import (
	"fmt"
	"github.com/spf13/viper"
	"io"
	"os"
	"strings"
	"time"
)

type IntRunArgs struct{ Start, Stop, Inc int }

type IntensityRun struct {
	Intensity, Worksize   IntRunArgs
	Runtime, AfterAllLock time.Duration
	AutoTune, Interleave  int
	ResetCards            bool
}

type InterleaveRun struct {
	Interleave            IntRunArgs
	Runtime, AfterAllLock time.Duration
	ResetCards            bool
}

func TuningRun(c *viper.Viper, run IntensityRun) error {
	switch {
	case run.Intensity.Start == 0 || run.Intensity.Stop == 0 || run.Intensity.Inc == 0:
		return fmt.Errorf("must provide valid intensity settings: start, stop and increment, %v %v %v", run.Intensity.Start, run.Intensity.Stop, run.Intensity.Inc)
	case run.Worksize.Start == 0 || run.Worksize.Stop == 0 || run.Worksize.Inc == 0:
		return fmt.Errorf("must provide worksize settings")
	}

	// err := ConfigLogger("logging.AmdConf", false)
	// if err != nil {
	// 	fmt.Printf("failed to configure logging")
	// }

	cards := NewCardData(c)
	cards.resetEnabled = run.ResetCards
	err := cards.GetStatus()
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
	debug(cards.String())
	if err = cards.ResetCards(false); err != nil {
		fmt.Printf("error Resetting cards %v\n", err)
	}

	api := NewStatsService(c).(*apiService)
	met := NewMetricsClient()
	met.enabled = true
	met.refresh = 10 * time.Second
	met.db = "tuningRun"

	if err = met.Config(c); err != nil {
		log.Infof("failed to config metrics client")
	}

	go met.backGroundWriter()

	go api.Monitor(met)

	dir := c.GetString("Core.Stak.Dir")
	if dir == "" {
		dir = "xmr-stak"
	}

	file := strings.Join([]string{dir, "amd.txt"}, pathSep)
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("Can't find File %v, %v", file, err)
	}

	amdConf := NewAmdConfig()
	if err = amdConf.Read(f); err != nil {
		log.Errorf("AmdConf.Read() error = %v", err)
	}
	f.Close()

	for wrksize := run.Worksize.Start; wrksize <= run.Worksize.Stop; wrksize += run.Worksize.Inc {
		if (run.Intensity.Start - run.Intensity.Start) > run.Intensity.Inc {
			fmt.Printf("invalid data start %v - stop %v <= increment %v\n", run.Intensity.Start, run.Intensity.Start, run.Intensity.Inc)
			continue
		}
		// align intensity to worksize
		intstart := highestDiv(run.Intensity.Start, wrksize)
		if err != nil {
			fmt.Printf("error calculating intensity %v", err)
			continue
		}
		intStop := highestDiv(run.Intensity.Stop, wrksize)
		inc := highestDiv(run.Intensity.Inc, wrksize)

		for i := intstart; i <= intStop; i += inc {
			// save amd.txt for run
			config := amdConf.GpuThreadsConf[0]
			config.Interleave = run.Interleave

			config.Intensity = i

			config.Worksize = wrksize

			for k := range amdConf.GpuThreadsConf {
				amdConf.GpuThreadsConf[k] = config
			}

			amdConf.AutoTune = run.AutoTune

			ms := tuneSession{
				confFile:      file,
				api:           api,
				ca:            cards,
				met:           met,
				amdConf:       amdConf,
				runTime:       run.Runtime,
				afterLock:     run.AfterAllLock,
				prMaxStartDur: c.GetDuration("Core.Stak.MaxProcessStartTime"),
			}
			// debug("intensity  %+v"+
			// 	" worksize %v"+
			// 	" interleave %v"+
			// 	" autoTune %v"+
			// 	" runtime %v", ms.intensity, ms.workSize, ms.interleave, ms.autoTune, ms.runTime)
			//
			err = RunMiner(ms, c)
			if err != nil {
				return fmt.Errorf("error mining %v", err)
			}
		}

		fmt.Println()
	}

	defer api.StopMonitor(met)

	return nil
}

type tuneSession struct {
	confFile                                  string
	api                                       *apiService
	ca                                        *CardData
	met                                       *metrics
	amdConf                                   AmdConf
	intensity, workSize, interleave, autoTune int
	runTime, afterLock                        time.Duration
	prMaxStartDur                             time.Duration
}

func RunMiner(s tuneSession, c *viper.Viper) error {
	gpuConf := s.amdConf.GpuThreadsConf
	if len(gpuConf) == 0 {
		return fmt.Errorf("no threads found in amd.txt")
	}

	err := s.ca.ResetCards(true)
	if err != nil {
		return fmt.Errorf("reset %v", err)
	}

	// config miner early so we can use kill method
	m := NewMiner()
	err = m.ConfigMiner(c)
	if err != nil {
		return fmt.Errorf("Failed configuring miner: %v\n", err)
	}

	err = m.killStak("RunMiner")
	if err != nil {
		log.Errorf("kilstak %v", err)
	}

	// write amd.conf to influx
	tags := map[string]string{"type": "AmdConf"}
	err = s.met.Write("config", tags, s.amdConf.Map())
	if err != nil {
		return fmt.Errorf("failed to write metrics %v", err)
	}
	err = s.met.Event(fmt.Sprintf("%+v", s.amdConf), "", "stak config")
	if err != nil {
		debug("failed to send event data %v", err)
	}

	// open conf file for writing
	f, err := os.OpenFile(s.confFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("failed to open %v for writing, got %v", s.confFile, err)
	}

	frwc := io.WriteCloser(f)

	// write config file
	err = s.amdConf.Write(frwc)
	if err != nil {
		return fmt.Errorf("failed to write %v got %v", s.confFile, err)
	}
	threadCount := len(s.amdConf.GpuThreadsConf)

	err = m.StartMining(s.prMaxStartDur)
	if err != nil {
		return fmt.Errorf("failed to start mining %v", err)
	}

	go m.ConsoleMetrics(s.met)

	err = s.api.tuningHash(s.runTime, s.afterLock, threadCount)
	if err != nil {
		return fmt.Errorf("tuningHash %v", err)
	}
	if err = m.StopMining("RunMiner()"); err != nil {
		return fmt.Errorf("failed to stop miner %v\n", err)
	}

	return nil
}

func InterleaveSession(c *viper.Viper, run InterleaveRun) error {
	// switch {
	// case run.Interleave.Start == 0 || run.Interleave.Stop == 0 || run.Interleave.Inc == 0:
	// 	return fmt.Errorf("must provide valid intensity settings: start, stop and increment, %v %v %v", run.Interleave.Start, run.Interleave.Stop, run.Interleave.Inc)
	// }

	cards := NewCardData(c)
	cards.resetEnabled = run.ResetCards
	err := cards.GetStatus()
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
	debug(cards.String())
	if err = cards.ResetCards(false); err != nil {
		fmt.Printf("error Resetting cards %v\n", err)
	}

	api := NewStatsService(c).(*apiService)
	met := NewMetricsClient()
	met.enabled = true
	met.refresh = 10 * time.Second
	met.db = "tuningRun"

	if err = met.Config(c); err != nil {
		log.Infof("failed to config metrics client")
	}

	go met.backGroundWriter()

	go api.Monitor(met)
	go api.ShowMonitor()

	dir := c.GetString("Core.Stak.Dir")
	if dir == "" {
		dir = "xmr-stak"
	}

	file := strings.Join([]string{dir, "amd.txt"}, pathSep)
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("Can't find File %v, %v", file, err)
	}

	amdConf := NewAmdConfig()
	if err = amdConf.Read(f); err != nil {
		log.Errorf("AmdConf.Read() error = %v", err)
	}
	f.Close()

	for interleave := run.Interleave.Start; interleave <= run.Interleave.Stop; interleave += run.Interleave.Inc {
		if (run.Interleave.Start - run.Interleave.Start) > run.Interleave.Inc {
			fmt.Printf("invalid data start %v - stop %v <= increment %v\n", run.Interleave.Start, run.Interleave.Start, run.Interleave.Inc)
			continue
		}
		// save amd.txt for run
		config := amdConf.GpuThreadsConf[0]
		config.Interleave = interleave

		for k := range amdConf.GpuThreadsConf {
			amdConf.GpuThreadsConf[k] = config
		}

		amdConf.AutoTune = 0

		ms := tuneSession{
			confFile:  file,
			api:       api,
			ca:        cards,
			met:       met,
			amdConf:   amdConf,
			runTime:   run.Runtime,
			afterLock: run.AfterAllLock,
		}
		fmt.Printf("interleave %v\n", interleave)
		err = RunMiner(ms, c)
		if err != nil {
			return fmt.Errorf("error mining %v", err)
		}
	}

	api.StopMonitor(met)

	return nil
}

func highestDiv(numerator, divider int) int {
	var result int
	if divider > numerator {
		return divider
	}
	result = (numerator / divider) * divider
	return result
}
