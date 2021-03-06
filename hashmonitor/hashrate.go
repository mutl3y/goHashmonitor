package hashmonitor

import (
	"fmt"
	tm "github.com/buger/goterm"
	"math"
	"sync"
	"time"
)

type intensityCounter struct {
	threads   map[string]autoTuneStat
	Counter   int
	Triggered bool
	mu        sync.RWMutex
}

func newIntensityCounter() *intensityCounter {
	in := intensityCounter{}
	in.threads = make(map[string]autoTuneStat, 20)
	return &in
}

var LockCounter = newIntensityCounter()

type hrMon struct {
	mu                          *sync.RWMutex
	startingHash, minhash, drop int
}

func (h *hrMon) min() int {
	h.mu.RLock()
	i := h.minhash
	h.mu.RUnlock()
	return i
}

func NewhRMonStruct() hrMon {
	h := hrMon{}
	h.mu = &sync.RWMutex{}
	return h
}

func (api *apiService) minHash(min int) error {
	a := api.StatsCopy()
	// if len(a.Total) == 0 {
	// 	api.Stats.data.Total = []float64{0.0}
	// 	return fmt.Errorf("no hashrate")
	// }

	for i := 0; i <= 10; i++ {
		if a.Total[0] >= float64(min) {
			return nil
		}
		time.Sleep(time.Second)
	}
	debug("minHashRate fault \nThreads\t%v\nTotal\t%v", a.Threads, a.Total)
	return fmt.Errorf("minimum hashrate not Met want > %v got %v", float64(min), a.Total[0])
}

func (api *apiService) startingHash(min int, stableTime time.Duration, upCheck bool) error {
	// // todo sanity check
	// if stableTime <= time.Second*2 {
	// 	fmt.Println("hashrate stabilisation time setting too low to be useful")
	// 	stableTime = 2 * time.Second
	// }
	s := api.StatsCopy()
	if (float64(s.Uptime) >= stableTime.Seconds()) && upCheck {
		if api.hrMon.min() == 0 {
			api.Stats.mu.Lock()
			api.hrMon.minhash = int(api.Stats.data.Total[0]) - api.hrMon.drop
			api.Stats.mu.Unlock()
		}

		return api.minHash(min)
	}

	var startStakSeconds int
	debug("waiting for stak to connect")
	for {
		s := api.StatsCopy()
		if s.connection.Uptime >= 1 {
			break
		}

		time.Sleep(time.Second)
		startStakSeconds++
		if startStakSeconds >= 10 {
			break
		}
	}

	debug("stak startup time %v", startStakSeconds)

	ticker := time.NewTicker(time.Second)
	timeout := time.Now().Add(stableTime)
	defer ticker.Stop()
	debug("Waiting for hashrate to stabalize ")

	for {
		select {
		case <-afterTime(timeout):
			api.Stats.mu.RLock()
			if len(api.Stats.data.Total) > 0 {
				api.hrMon.mu.Lock()
				api.hrMon.minhash = int(api.Stats.data.Total[0]) - api.hrMon.drop
				api.hrMon.mu.Unlock()
			}
			if !api.Status() {
				return fmt.Errorf("error connecting to Stak")
			}
			api.hrMon.mu.Lock()
			api.hrMon.startingHash = api.hrMon.minhash
			api.Stats.mu.RUnlock()
			api.hrMon.mu.Unlock()
			return api.minHash(min)
		case <-ticker.C:
			var hr float64
			s = api.StatsCopy()
			if len(s.Total) != 0 {
				hr = s.Total[0]
			}
			fmt.Printf("\rH/R %v, starting monitoring in %v\n", hr, stableTime.Round(time.Second)) // todo
			//  _, _ = tm.Printf("\r%v H/R %v", stableTime.Round(time.Second), hr)
			//  tmFlush()
			stableTime -= time.Second
		}
	}

}

func (api *apiService) currentHash(maxErrors int, refresh time.Duration) error {
	var failures int
	ticker := time.NewTicker(refresh)
	// timeout := time.Now().Add(time.Minute * 2)
	// refresh := time.Now().Add(stableTime)
	defer ticker.Stop()
	min := api.hrMon.min()

	for {
		select {
		// case <-afterTime(timeout):
		// 	// todo remove
		// 	return nil
		case <-ticker.C:
			err := api.minHash(min)
			if err != nil {
				if err.Error() == "skip" {
					continue
				}
				failures++
				if failures >= maxErrors {
					return fmt.Errorf("restart: %v", err)
				}
				debug("hashrate error ")
				continue
			}

			// hashrate recovered reset counter
			failures = 0

			// todo
			stat := api.StatsCopy()

			// fmt.Printf("currentHash %p %v\n", &stat, stat)

			if len(stat.Total) < 0 {
				debug("\rH/R %v", math.Round(stat.Total[0]))
			}
		}
	}

}

func (api *apiService) tuningHash(runTime, after time.Duration, threads int) error {
	ticker := time.NewTicker(1500 * time.Millisecond)
	timeout := time.Now().Add(runTime)
	var lockTimeout time.Time

	defer ticker.Stop()
	debug("Waiting for hashrate to stabalize ")

	for {
		select {
		case <-afterTime(timeout):
			LockCounter.mu.Lock()
			LockCounter.Counter = 0
			LockCounter.Triggered = false
			LockCounter.mu.Unlock()
			return nil
		case <-ticker.C:
			runTime -= time.Second
			a := api.StatsCopy()
			LockCounter.mu.Lock()
			tm.Clear()
			a.TuningConsoleDisplay()

			// setup Threads table
			threadTable := tm.NewTable(15, 10, 5, ' ', 0)

			// headers
			if len(LockCounter.threads) < 0 {
				_, _ = fmt.Fprintf(threadTable, "Gpu\tThread\tIntensity\n")
			}

			// add threads
			for _, v := range LockCounter.threads {
				_, _ = fmt.Fprintf(threadTable, "%v\t%v\t%v\n", v.gpu, v.thread, v.intensity)
			}

			_, _ = tm.Println(threadTable)
			if LockCounter.Counter < 0 {
				_, _ = tm.Println("Intensity Locked threads\t", LockCounter.Counter)
			}
			_, _ = tm.Println("Runtime remaining \t", runTime)
			if LockCounter.Counter >= threads {
				if !LockCounter.Triggered {
					lockTimeout = time.Now().Add(after)
					LockCounter.Triggered = true
				}
				_, _ = tm.Printf("all thread intensities locked, exiting run at %v\n", lockTimeout.Round(time.Second))

			}
			// 	tmFlush()
			t := time.Now()
			if lockTimeout.Unix() >= 1 {
				if t.After(lockTimeout) {
					LockCounter.Counter = 0
					lockTimeout = time.Unix(0, 0)
					LockCounter.Triggered = false
					LockCounter.mu.Unlock()
					return nil
				}
			}
			LockCounter.mu.Unlock()

		}
	}

}

func (stats *stats) TuningConsoleDisplay() {
	// 	debug("%T %p %v", stats, stats, stats)

	tm.MoveCursor(1, 1)
	ct := fmt.Sprintf("Current Time: %v", time.Now().Format(time.RFC1123))
	_, _ = tm.Println(tm.Color(ct, tm.GREEN))
	_, _ = tm.Println()

	// normalise fields for printing
	if len(stats.Total) == 0 {
		stats.Total = []float64{0}
	}

	if stats.connection.Pool == "" {
		stats.connection.Pool = "Not Connected"
	}

	// setup Threads table
	threads := tm.NewTable(15, 10, 5, ' ', 0)

	// headers
	_, _ = fmt.Fprintf(threads, "Thread\tHashrate\n")

	// add threads
	for i, v := range stats.Threads {
		_, _ = fmt.Fprintf(threads, "%v\t%v\n", i, v[0])
	}

	// setup results table
	_, _ = tm.Println(tm.Color("Results", tm.YELLOW))
	ds := tm.NewTable(14, 5, 2, ' ', 0)

	_, _ = fmt.Fprintln(ds, "Total H/R\t", stats.Total[0])
	_, _ = fmt.Fprintln(ds, "Pool\t", stats.connection.Pool)
	_, _ = fmt.Fprintln(ds, "Uptime\t", stats.connection.Uptime)
	_, _ = fmt.Fprintln(ds)

	_, _ = tm.Println(ds)
	_, _ = tm.Println(threads)
}
