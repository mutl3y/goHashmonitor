package hashmonitor

import (
	"fmt"
	tm "github.com/buger/goterm"
	"math"
	"time"
)

func (api *apiService) minHash(min int) error {
	a := api.StatsCopy()
	if len(a.Total) == 0 {
		api.Stats.mu.Lock()
		api.Stats.data.Total = []float64{0.0}
		api.Stats.mu.Unlock()
		return nil
	}

	if a.Total[0] < float64(min) {
		return fmt.Errorf("minimum hashrate not met")
	}
	return nil
}

func (api *apiService) startingHash(min int, stableTime time.Duration) error {
	// // todo sanity check
	// if stableTime <= time.Second*2 {
	// 	fmt.Println("hashrate stabilisation time setting too low to be useful")
	// 	stableTime = 2 * time.Second
	// }
	s := api.StatsCopy()

	if float64(s.Uptime) >= stableTime.Seconds() {
		log.Debugf("Stak already up")
		return api.minHash(min)
	}

	ticker := time.NewTicker(time.Second)
	timeout := time.Now().Add(stableTime)
	defer ticker.Stop()
	debug("Waiting for hashrate to stabalize ")

	for {
		select {
		case <-afterTime(timeout):
			return api.minHash(min)
		case <-ticker.C:
			var hr float64
			s = api.StatsCopy()
			if len(s.Total) != 0 {
				hr = s.Total[0]
			}
			tm.Clear()
			_, _ = tm.Printf("\r%v H/R %v", stableTime.Round(time.Second), hr)
			tm.Flush()
			stableTime -= time.Second
		}
	}

}

func (api *apiService) currentHash(hash, maxErrors int, refresh time.Duration) error {
	var failures int
	ticker := time.NewTicker(refresh)
	timeout := time.Now().Add(time.Minute * 10)
	// refresh := time.Now().Add(stableTime)
	defer ticker.Stop()

	for {
		select {
		case <-afterTime(timeout):
			// todo remove
			return nil
		case <-ticker.C:
			if err := api.minHash(hash); err != nil {
				if err.Error() == "skip" {
					continue
				}
				failures++
				if failures <= maxErrors {
					return fmt.Errorf("hashrate has dropped")
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
