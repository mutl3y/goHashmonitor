package hashmonitor

import (
	"encoding/json"
	"fmt"
	tm "github.com/buger/goterm"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ApiService implementation interface
type ApiService interface {
	Monitor(met *metrics) bool
	stopMonitor(met *metrics) bool
	showMonitor()
}
type apiService struct {
	URL    string
	Signal chan bool
	limit  *simpleRateLimit
	Stats  *rwStats
	Up     bool
}

// NewStatsService returns a monitoring service with rate limiter
// takes settings from viper config
func NewStatsService(cfg *viper.Viper) ApiService {
	Signal := make(chan bool)
	apiIp := cfg.GetString("Core.Stak.Ip")
	apiPort := cfg.GetInt64("Core.Stak.Port")
	apiUrl := fmt.Sprintf("http://%v:%v/api.json", apiIp, apiPort)

	debug("NewStatsService, %v:%v ", apiIp, apiPort)

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
		Stats:  new(rwStats),
	}
}

func (api *apiService) StatsCopy() stats {

	api.Stats.mu.RLock()

	stat := stats{}
	// stat = api.Stats.data
	fmt.Printf("statscopy1 %T %p %v\n", stat.Total, &stat.Total, stat.Total)
	fmt.Printf("statscopy2 %T %p %v\n", api.Stats.data.Total, &api.Stats.data.Total, api.Stats.data.Total)
	fmt.Printf("api     %+v\n", api.Stats.data)
	fmt.Printf("stat    %+v\n", stat)

	api.Stats.mu.RUnlock()
	return stat
}

func (api *apiService) StatsUpdate(s stats) {
	debug("su %v", *api)
	api.Stats.mu.Lock()
	defer api.Stats.mu.Unlock()

	api.Stats.data.LastUpdate = time.Now()
	api.Stats.data = s

	return
}

// Monitor Starts monitoring Stak
func (api *apiService) Monitor(m *metrics) bool {
	errChan := make(chan error, 10)
	go func(err chan error, api *apiService) {
		timeout := time.Duration(500 * time.Millisecond)
		client := http.Client{
			Timeout: timeout,
		}
		timeoutError := "request canceled"

		for api.Signal != nil {
			select {
			case <-api.Signal:
				api.limit.Signal <- true
				errChan <- fmt.Errorf("capiService.Monitor signal channel closed")
				close(errChan)
				return
			case <-api.limit.throttle:

				res, err := client.Get(api.URL)
				if err != nil {
					if strings.Contains(err.Error(), timeoutError) {
						continue
					}
					errChan <- fmt.Errorf("error connecting: %v", err)
					api.Up = false
					debug("%v", err)
					continue
				}
				if res.StatusCode != 200 {
					errChan <- fmt.Errorf("%v", res.Status)
					continue
				}
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					errChan <- fmt.Errorf("error reading Body: %v", err)
					continue
				}

				out := stats{}
				err = json.Unmarshal(body, &out)
				if err != nil {
					fmt.Println(err)
					debug("error unmarshaling JSON: %v", err)
					continue
				}
				err = res.Body.Close()
				if err != nil {
					log.Fatalf("failed to close body, stopping under the no leak policy %v", err)
				}

				api.StatsUpdate(out)
				api.Up = true

				if met := out.Map(); err == nil {
					hostname, herr := os.Hostname()
					if herr != nil {
						log.Errorf("failed to set hostname %v", herr)
					}
					tags := map[string]string{"server": hostname}
					err = m.Write("metrics", tags, met)
					if err != nil {
						debug("failed to write to influx %v", err)
					}
				}

			case outerr := <-errChan:
				log.Errorf("%v", outerr)
			}
		}
	}(errChan, api)

	return true
}

func (api *apiService) stopMonitor(m *metrics) bool {
	api.Signal <- true
	debug("stopping Stats Service")
	m.Stop()
	return true
}

func (api *apiService) showMonitor() {
	limit := newLimiter(500 * time.Millisecond)
	go limitClock(limit)

	for {
		select {
		case <-limit.throttle:
			a := api.StatsCopy()
			if api.Up {
				a.ConsoleDisplay()
			}
		case <-limit.Signal:
			return
		case _, ok := <-api.Signal:
			if !ok {
				if !limit.Stop() {
					debug("limiter didn't stop")
				}
				return
			}
		}
	}
}

// stak api data structs
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
	LastUpdate time.Time `json:"last,omitempty"`
}

// type noCopy struct{}
//
// func (*noCopy) Lock() {}
// type Locker interface {
// 	Lock()
// 	Unlock()
// }
type rwStats struct {
	mu   sync.RWMutex
	data stats
}

// Map returns a map version of stats data for metrics.go
// non concurrent usage
func (stats *stats) Map() map[string]interface{} {
	m := map[string]interface{}{
		"DiffCurrent": stats.DiffCurrent,
		"SharesGood":  stats.SharesGood,
		"SharesTotal": stats.SharesTotal,
		"AvgTime":     stats.AvgTime,
		"HashesTotal": stats.HashesTotal,
		"Pool":        stats.Pool,
		"Uptime":      stats.Uptime,
		"Ping":        stats.Ping,
	}

	for k, v := range stats.Threads {
		m[fmt.Sprintf("Thread_%v", k)] = v[0]
	}
	return m
}
func (stats *stats) ConsoleDisplay() {
	tm.Clear()
	tm.MoveCursor(1, 1)
	_, _ = tm.Println("Current Time:", time.Now().Format(time.RFC1123))
	for _, v := range stats.Threads {
		_, _ = tm.Printf("%s %+v\t", "\033[H\033[2J", v[0])

	}

	tm.Flush() // Call it every time at the end of rendering
}

func simApi(api *apiService, wg *sync.WaitGroup, startHashRate int, decayRate float64, decayTime time.Duration) {
	ticker := time.NewTicker(decayTime)
	defer ticker.Stop()
	timeout := time.Now().Add(time.Second * 30)
	// refresh := time.Now().Add(stableTime)
	// api.Lock()
	stat := stats{}
	stat.Total = []float64{float64(startHashRate)}

	api.StatsUpdate(stat)
	// api.Unlock()
	wg.Done()
	for {
		select {
		case <-afterTime(timeout):
			return
		case <-ticker.C:
			st := api.StatsCopy()
			// 	stat.Total = st.Total
			// fmt.Printf("simApi    %p %v\n", &stat, stat)
			fmt.Printf("simapi     %T %p %v\n", stat.Total[0], &stat.Total[0], stat.Total[0])
			fmt.Printf("stp        %T %p %v\n", st.Total, &st.Total, st.Total)

			if len(st.Total) == 0 {
				st.Total = []float64{0.0}
			}

			stat.Total[0] = st.Total[0] / decayRate
			api.StatsUpdate(stat)
			if stat.Total[0] <= 10 {
				return
			}
		}
	}

}
