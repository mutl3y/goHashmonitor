package hashmonitor

import (
	"fmt"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type MineSession struct {
	confFile, intCheckUrl                              string
	Api                                                *apiService
	ca                                                 *CardData
	Met                                                *metrics
	m                                                  *miner
	AmdConf                                            AmdConf
	startFailures, minHashRate, maxErrors, restartWait int
	stableTime, refreshTime, intHttpTimeout            time.Duration
	resetEnabled                                       bool
}

func NewMineSession(v *viper.Viper) (*MineSession, error) {
	ms := MineSession{}
	ms.Api = NewStatsService(v).(*apiService)
	ms.m = NewMiner()
	err := ms.m.ConfigMiner(v)
	if err != nil {
		return &MineSession{}, fmt.Errorf("newMinningService() failed to configure miner %v", err)
	}
	ms.Met = NewMetricsClient()
	ms.startFailures = v.GetInt("Core.Stak.Start_Attempts")
	ms.minHashRate = v.GetInt("Core.Hash.Min")
	ms.stableTime = v.GetDuration("Core.Stak.Stable_Time")
	if ms.stableTime <= 20*time.Second {
		ms.stableTime = 20 * time.Second
	}

	ms.refreshTime = v.GetDuration("Core.Stak.Refresh_Time")
	ms.ca = NewCardData(v)
	dir := v.GetString("Core.Stak.Dir")
	if dir == "" {
		dir = "xmr-stak"
	}

	file := strings.Join([]string{dir, "amd.txt"}, string(os.PathSeparator))
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("Can't find File %v, %v", file, err)
	}

	ms.AmdConf = NewAmdConfig()
	if err = ms.AmdConf.Read(f); err != nil {
		log.Fatalf("AmdConf.Read() error = %v", err)
	}
	f.Close()

	ms.intCheckUrl = v.GetString("Core.Connection.Check.Destination")
	ms.intHttpTimeout = v.GetDuration("Core.Connection.Check.Seconds")

	return &ms, nil
}

func (ms *MineSession) Mine() error {
	gpuConf := ms.AmdConf.GpuThreadsConf

	if len(gpuConf) == 0 {
		return fmt.Errorf("no threads found in amd.txt")
	}

	err := ms.ca.ResetCards(ms.ca.resetOnStartUp)
	if err != nil {
		return fmt.Errorf("reset %v", err)
	}

	// err = ms.m.killStak()
	// if err != nil {
	// 	log.Errorf("killStak %v", err)
	// }

	// write amd.conf to influx
	tags := map[string]string{"type": "AmdConf"}
	err = ms.Met.Write("config", tags, ms.AmdConf.Map())
	if err != nil {
		return fmt.Errorf("failed to write metrics %v", err)
	}
	err = ms.Met.Event(fmt.Sprintf("%+v", ms.AmdConf), "", "stak config")
	if err != nil {
		debug("failed to send event data %v", err)
	}

	ms.Api.Monitor(ms.Met)
	defer ms.Api.StopMonitor(ms.Met)
	if !Debug {
		go ms.Api.ShowMonitor()
	}

	err = ms.m.RunTools()
	if err != nil {
		log.Errorf("%v", err)
	}
	err = ms.MiningSession(0)
	if err != nil {
		log.Errorf("mining session error %v", err)
		if ms.resetEnabled {
			err = RestartCommputer(ms.restartWait)
			if err != nil {
				log.Errorf("failed to restart computer %v", err)
			}
		}
	}

	return err
}

func checkInternet(url string, timeout time.Duration, maxFails int) error {
	var failsThisCheck int
	if timeout <= 2*time.Second {
		timeout = 2 * time.Second
	}

	for {
		client := http.Client{
			Timeout: timeout,
		}
		timeoutError := "request canceled"
		_, err := client.Get("http://" + url)
		if err == nil {
			return nil
		}
		if err.Error() != timeoutError {
			return err

		}

		failsThisCheck++
		if failsThisCheck >= maxFails {
			return fmt.Errorf("max fails exhausted")
		}
	}
}

func (ms *MineSession) MiningSession(maxFail int) (err error) {
	var startFailures, failures, procMissing int

	if maxFail == 0 {
		maxFail = 99965535
	}

	debug("new mining session")
	defer func() {
		err = ms.m.killStak("Miningsession Defer")
	}()

	start := func() error {
		// stop console metrics if running in the background
		// 	ms.m.SetUp(false)
		err = ms.m.StartMining()
		if err != nil {
			return fmt.Errorf("failed to start mining %v", err)
		}

		go ms.m.ConsoleMetrics(ms.Met)
		fmt.Println("Pausing for 5 seconds to allow Stak to start")
		time.Sleep(5 * time.Second)

		return ms.m.CheckStakProcess()
	}

	for failures <= maxFail {
		if procMissing >= 2 {
			return fmt.Errorf("stak terminating abnormally")
		}
		debug("start failures \t%v \tMonitoring failures \t%v", startFailures, failures)

		if e := checkInternet(ms.intCheckUrl, ms.intHttpTimeout, maxFail); e != nil {
			return e
		}

		// check for a stak process and start a new miner session if not found
		e := ms.m.CheckStakProcess()
		if e != nil {
			switch e.Error() {

			case "process does not exist":
				procMissing++
				fallthrough
			case "no process being tracked":
				if e := start(); e != nil {
					return e
				}
			default:
				return fmt.Errorf("ms.CheckStakProcess() error %v", e)

			}

		}

		// check stak Api and start a new miner session if not responsive
		if e := ms.Api.CheckApi(4, time.Second); e != nil {
			_ = ms.m.killStak("miningsession() checkApi")
			if e := start(); e != nil {
				return e
			}
		}

		// let Stak settle before monitoring starts, skip if already running
		loopErr := ms.Api.startingHash(ms.minHashRate, ms.stableTime, true)
		if loopErr != nil {
			startFailures++
			if startFailures >= ms.startFailures {
				return fmt.Errorf("MiningSession startingHash %v", loopErr)
			} else {
				log.Errorf("start failure %v", loopErr)
				continue
			}
		} else {
			// reset bad start count
			startFailures = 0
		}

		// monitor running stak
		if loopErr := ms.Api.currentHash(10, ms.refreshTime); loopErr != nil {
			log.Infof("restarting monitoring %v", loopErr)
			failures++
		}

	}
	return nil
}

func RestartCommputer(t int) (err error) {

	switch Os := runtime.GOOS; {
	case Os == "windows":
		cmd := exec.Command("shutdown", "/r", "/t", string(t))
		err = cmd.Run()
		if err != nil {
			return err
		}

		log.Error("restarting computer ")

	case Os == "linux":
		cmd := exec.Command("shutdown", "-r", "now")
		err = cmd.Run()
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("%v not supported", Os)
	}

	return nil
}

func CancelRestart(t string) (err error) {

	switch Os := runtime.GOOS; {
	case Os == "windows":
		cmd := exec.Command("shutdown", "/a")
		err = cmd.Run()
		if err != nil {
			return err
		}

		fmt.Printf("restart computer ")

	case Os == "linux":
		cmd := exec.Command("shutdown", "-r", "now")
		err = cmd.Run()
		if err != nil {
			return err
		}

		fmt.Printf("restart computer ")

	default:
		return fmt.Errorf("%v not supported", Os)
	}

	return nil
}
