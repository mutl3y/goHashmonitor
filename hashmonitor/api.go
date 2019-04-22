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

type errorLog struct {
	Count    int    `json:"count"`
	LastSeen int    `json:"last_seen"`
	Text     string `json:"text"`
}

type hashrate struct {
	Threads [][]float64 `json:"threads"`
	Total   []float64   `json:"total"`
	Highest float64     `json:"highest"`
}
type results struct {
	DiffCurrent int        `json:"diff_current"`
	SharesGood  int        `json:"shares_good"`
	SharesTotal int        `json:"shares_total"`
	AvgTime     float64    `json:"avg_time"`
	HashesTotal int        `json:"hashes_total"`
	Best        []int      `json:"best"`
	ErrorLog    []errorLog `json:"error_log"`
}
type connection struct {
	Pool     string        `json:"pool"`
	Uptime   int           `json:"uptime"`
	Ping     int           `json:"ping"`
	ErrorLog []interface{} `json:"error_log"`
}

type stats struct {
	Version    string `json:"version"`
	hashrate   `json:"hashrate"`
	results    `json:"results"`
	connection `json:"connection"`
}

func newStats() stats {
	s := new(stats)
	s.hashrate.Threads = make([][]float64, 0, 20)
	s.hashrate.Total = make([]float64, 0, 20)
	s.results.Best = make([]int, 0, 1)
	s.results.ErrorLog = make([]errorLog, 0, 10)
	return *s
}

func (s *stats) Map() (map[string]interface{}, error) {
	m := map[string]interface{}{
		"DiffCurrent": s.results.DiffCurrent,
		"SharesGood":  s.results.SharesGood,
		"SharesTotal": s.results.SharesTotal,
		"AvgTime":     s.results.AvgTime,
		"HashesTotal": s.results.HashesTotal,
		"Pool":        s.connection.Pool,
		"Uptime":      s.connection.Uptime,
		"Ping":        s.connection.Ping,
	}
	threads := s.hashrate.Threads
	for k, v := range threads {
		m[fmt.Sprintf("thread_%v", k)] = v[0]
	}
	return m, nil
}

type stakStats struct {
	sync.RWMutex
	data       stats
	LastUpdate time.Time
}

//type signal chan struct{}

type ApiService interface {
	Monitor() bool
	StopMonitor() bool
	ShowMonitor() error
}

type simpleRateLimit struct {
	Signal   chan bool
	throttle chan time.Time
	rate     time.Duration
}

type apiService struct {
	URL    string
	Signal chan bool
	limit  *simpleRateLimit
	Stats  *stakStats
	Up     bool
	//	Connected bool
}

// Monitor Starts monitoring Stak
func (api *apiService) Monitor() bool {
	errChan := make(chan error, 10)
	go func(err chan error, api *apiService) {
		m := NewMetricsClient()
		if err := m.Config(cfg); err != nil {
			log.Infof("failed to config metrics client")
		}
		if err2 := m.checkDB(); err2 == nil {
			go m.backGroundWriter()
			defer close(m.pointsQueue)
		}

	loop:
		for {
			select {
			case _, ok := <-api.Signal:
				if !ok {
					errChan <- fmt.Errorf("channel closed")
					close(errChan)
					api.limit.Signal <- true
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
					fmt.Println(err)
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
				//	debug("api connected %v",api.Connected)
				api.Stats.Unlock()

				if met, err := out.Map(); err == nil {
					tags := map[string]string{"server": "hostnamegoeshere"}
					m.Write("metrics", tags, met)
				}

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
	defer func() { limit.Signal <- true }()
	x := 0
loop:
	for {

		select {
		case <-limit.throttle:
			if api.Up {
				api.Stats.RLock()
				fmt.Printf("%s %+v\n", "\033[H\033[2J", api.Stats.data)
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
	Signal := make(chan bool)
	apiIp := cfg.GetString("Core.Stak.Ip")
	apiPort := cfg.GetInt64("Core.Stak.Port")
	apiUrl := fmt.Sprintf("http://%v:%v/api.json", apiIp, apiPort)

	log.Debugf("NewStatsService, %v:%v ", apiIp, apiPort)

	// Start a refresh limiter
	minlimit := time.Millisecond * 500
	l := cfg.GetDuration("Core.Stak.refresh_ms")
	if l < minlimit {
		l = minlimit
	}

	limit := newLimiter(l)

	go limitClock(limit)
	return &apiService{
		URL:    apiUrl,
		Signal: Signal,
		limit:  limit,
		Stats:  new(stakStats),
	}
}

// newLimiter Takes a time duration for refresh speed
// returns a simple rate limiter Config with a signal channel
// uses select so non blocking
func newLimiter(rate time.Duration) *simpleRateLimit {
	t := make(chan time.Time, 1)
	c := make(chan bool)
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
			log.Debugf("stopping rate limiter")
			return
		}
	}
}
