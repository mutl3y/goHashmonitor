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
	resetEnabled, resetFailedStart                     bool
	prMaxStartDur                                      time.Duration
}

func NewMineSession(v *viper.Viper) (*MineSession, error) {
	ms := MineSession{}
	ms.Api = NewStatsService(v).(*apiService)

	return &ms, nil
}

func (ms *MineSession) Config(v *viper.Viper) error {
	debug("configuring miner")
	ms.m = NewMiner()
	err := ms.m.ConfigMiner(v)
	if err != nil {
		return fmt.Errorf("newMinningService() failed to configure miner %v", err)
	}
	ms.Met = NewMetricsClient()
	ms.startFailures = v.GetInt("Core.Stak.Start_Attempts")
	ms.minHashRate = v.GetInt("Core.Hash.Min")
	ms.stableTime = v.GetDuration("Core.Stak.Stable_Time")

	// next 2 lines should be 20
	if ms.stableTime <= 2*time.Second {
		ms.stableTime = 2 * time.Second
	}
	ms.prMaxStartDur = v.GetDuration("Core.Stak.MaxProcessStartTime")
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
	ms.resetFailedStart = v.GetBool("Core.Stak.StartFailure.Reboot")

	return nil
}

func (ms *MineSession) Mine() error {
	gpuConf := ms.AmdConf.GpuThreadsConf
	if len(gpuConf) == 0 {
		return fmt.Errorf("no threads found in amd.txt")
	}

	// err = ms.m.killStak()
	// if err != nil {
	// 	log.Errorf("killStak %v", err)
	// }

	// write amd.conf to influx
	tags := map[string]string{"type": "AmdConf"}
	err := ms.Met.Write("config", tags, ms.AmdConf.Map())
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

	// main mining loop logic goes here
	for c := 1; c <= 20; c++ {
		fmt.Println("run", c) // todo
		err := ms.ca.ResetCards(ms.ca.resetOnStartUp)
		if err != nil {
			return fmt.Errorf("reset %v", err)
		}

		err = ms.m.RunTools()
		if err != nil {
			log.Errorf("%v", err)
		}

		err = ms.MiningSession(2)
		if err != nil {
			log.Errorf("%v", err)
			estring := err.Error()

			switch {
			case strings.HasPrefix(estring, "start:"):
				if ms.resetFailedStart {
					err = RestartCommputer(ms.restartWait)
					if err != nil {
						log.Errorf("failed to restart computer %v", err)
					}
				} else {
					log.Errorf("need a reboot but restarts are not enabled")
				}
				return fmt.Errorf("start issue %v", err)
			case strings.HasPrefix(estring, "restart:"):
				ms.Api.hrMon.mu.Lock()
				ms.Api.hrMon.startingHash = 0
				ms.Api.hrMon.mu.Unlock()
				ms.ca.resetOnStartUp = true
			case strings.HasPrefix(estring, "procMissing:"):
				return fmt.Errorf("stak terminating abnormally please investigate")
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
		if !strings.Contains(err.Error(), timeoutError) {
			return err

		}

		failsThisCheck++
		if failsThisCheck >= maxFails {
			return fmt.Errorf("max fails exhausted")
		}
	}
}

func (ms *MineSession) mineOnce(checks int, checkTime time.Duration) (err error) {

	err = ms.Api.CheckApi(checks, checkTime)
	if err != nil {
		_ = ms.m.killStak("mineOnce()")
		err = ms.m.StartMining(ms.prMaxStartDur)
		if err != nil {
			return fmt.Errorf("failed to start mining %v", err)
		}

		go ms.m.ConsoleMetrics(ms.Met)
	}

	return err
}

func (ms *MineSession) MiningSession(maxFail int) error {
	var startFailures, failures, procMissing int
	var errString string
	if maxFail == 0 {
		maxFail = 99965535
	}

	debug("new mining session")

	// lets make sure Stak exits
	defer func() {
		_ = ms.m.killStak("Miningsession Defer")
	}()

	for failures < maxFail {
		if procMissing >= 2 { //
			return fmt.Errorf("procMissing: ")
		}

		if e := checkInternet(ms.intCheckUrl, ms.intHttpTimeout, maxFail); e != nil {
			return e
		}

		e := ms.m.CheckStakProcess()
		if e != nil {
			switch e.Error() {
			case "process does not exist":
				procMissing++
				fallthrough
			case "no process being tracked":
				if e := ms.mineOnce(10, 500*time.Millisecond); e != nil {
					return e
				}
				continue
			default:
				return fmt.Errorf("checkStakProcess: %v", e)

			}

		}
		procMissing = 0
		debug("Start failures \t%v \tMonitoring failures \t%v", startFailures, failures)

		loopErr := ms.Api.startingHash(ms.minHashRate, ms.stableTime, true)
		if loopErr != nil {
			// Stop the console being updated and reset the starting hash for the next run
			ms.Api.hrMon.mu.Lock()
			ms.Api.hrMon.startingHash = 0
			ms.Api.hrMon.mu.Unlock()

			if err := ms.m.killStak("startingHash"); err != nil {
				debug("error killing miner")
			}

			startFailures++
			if startFailures >= ms.startFailures {

				return fmt.Errorf("start: %v", loopErr)
			} else {
				debug("start failure %v", loopErr)
				time.Sleep(time.Second)
				continue
			}
		}

		// reset bad startFailures count, if we got this far it started
		startFailures = 0

		// monitor running stak
		if err := ms.Api.currentHash(ms.maxErrors, ms.refreshTime); err != nil {
			debug("%v", err)
			failures++
			errString = err.Error()
		} else {
			failures = 0
		}

	}

	return fmt.Errorf("%v", errString)
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
