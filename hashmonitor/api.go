package hashmonitor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type stakStats struct {
	sync.RWMutex
	data struct {
		Version  string `json:"version"`
		Hashrate struct {
			Threads [][]float64 `json:"threads"`
			Total   []float64   `json:"total"`
			Highest float64     `json:"highest"`
		} `json:"hashrate"`
		Results struct {
			DiffCurrent int     `json:"diff_current"`
			SharesGood  int     `json:"shares_good"`
			SharesTotal int     `json:"shares_total"`
			AvgTime     float64 `json:"avg_time"`
			HashesTotal int     `json:"hashes_total"`
			Best        []int   `json:"best"`
			ErrorLog    []struct {
				Count    int    `json:"count"`
				LastSeen int    `json:"last_seen"`
				Text     string `json:"text"`
			} `json:"error_log"`
		} `json:"results"`
		Connection struct {
			Pool     string        `json:"pool"`
			Uptime   int           `json:"uptime"`
			Ping     int           `json:"ping"`
			ErrorLog []interface{} `json:"error_log"`
		} `json:"connection"`
	}
	LastUpdate time.Time
}
type signal chan struct{}

type ApiService interface {
	Monitor() bool
	StopMonitor() bool
	ShowMonitor() error
}

type simpleRateLimit struct {
	Signal   chan struct{}
	throttle chan time.Time
	rate     time.Duration
}

type apiService struct {
	URL       string
	Signal    chan struct{}
	limit     *simpleRateLimit
	Stats     *stakStats
	Up        bool
	Connected bool
}

// Monitor Starts monitoring Stak
func (api *apiService) Monitor() bool {
	errChan := make(chan error, 10)
	go func(err chan error, api *apiService) {
	loop:
		for {
			select {
			case _, ok := <-api.Signal:
				if !ok {
					errChan <- fmt.Errorf("channel closed")
					close(errChan)
					break loop

				}
			case <-api.limit.throttle:
				res, err := http.Get(api.URL)
				if err != nil {
					errChan <- fmt.Errorf("error connecting: %v", err)
					api.Up = false
					log.Debugf("%v", err)
					return
				}
				if res.StatusCode != 200 {
					errChan <- fmt.Errorf("%v", res.Status)
					return
				}
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					errChan <- fmt.Errorf("error reading Body: %v", err)
					return
				}

				out := stakStats{}.data
				err = json.Unmarshal(body, &out)
				if err != nil {
					errChan <- fmt.Errorf("error unmarshaling JSON: %v", err)
					return
				}
				err = res.Body.Close()
				if err != nil {
					log.Fatalf("failed to close body %v", err)
				}

				api.Stats.Lock()
				api.Stats.data = out
				api.Stats.LastUpdate = time.Now()
				api.Up = true
				fmt.Println(api.Connected)
				api.Stats.Unlock()

			case outerr := <-errChan:
				log.Errorf("%v", outerr)
			}
		}
	}(errChan, api)

	return true
}

func (api *apiService) StopMonitor() bool {
	close(api.Signal)
	close(api.limit.Signal)
	if _, ok := <-api.Signal; !ok {
		log.Debugf("Api Signaled, Exiting \n")
	}
	if _, ok := <-api.limit.Signal; !ok {
		log.Debugf("Limiter Signaled, Exiting \n")
	}
	return true
}

func (api *apiService) ShowMonitor() error {
	limit := newLimiter(500 * time.Millisecond)
	go limitClock(limit)
	defer func() { limit.Signal <- struct{}{} }()
	x := 0
loop:
	for {

		select {
		case <-limit.throttle:
			if api.Connected {
				api.Stats.RLock()
				fmt.Printf("%s %+v\n", "\033[H\033[2J", api.Stats.LastUpdate)
				api.Stats.RUnlock()
			}
		case <-limit.Signal:
			return errors.New("stopped")
		case <-time.After(15 * time.Second):
			return errors.New("timed out")

		}
		x++
		if x >= 10 {
			break loop
		} // todo
	}

	return nil
}

func NewStatsService(cfg *viper.Viper) ApiService {
	Signal := make(chan struct{})
	apiIp := cfg.Get("Core.Stak.Ip").(string)
	apiPort := cfg.Get("Core.Stak.Port").(int)
	apiUrl := fmt.Sprintf("http://%v:%v/api.json", apiIp, apiPort)
	stakStats := new(stakStats)

	// Start a refresh limiter
	limit := newLimiter(cfg.GetDuration("Core.Stak.refresh_ms"))
	go limitClock(limit)
	return &apiService{
		URL:    apiUrl,
		Signal: Signal,
		limit:  limit,
		Stats:  stakStats,
	}
}

// newLimiter Takes a time duration for refresh speed
// returns a simple rate limiter Config with a signal channel
// uses select so non blocking
func newLimiter(rate time.Duration) *simpleRateLimit {
	t := make(chan time.Time, 1)
	c := make(chan struct{})
	limit := simpleRateLimit{Signal: c, throttle: t, rate: rate}
	return &limit
}

func limitClock(limit *simpleRateLimit) {
	tick := time.NewTicker(limit.rate)
	defer tick.Stop()
	for t := range tick.C {
		select {
		case limit.throttle <- t:
		case <-limit.Signal:
			return
		}
	}
}
