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

var safeFlush = &sync.Mutex{}

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

	stat := &stats{}
	*stat = api.Stats.data
	// fmt.Printf("api %T %p %v\n", api.Stats.data.Total, &api.Stats.data.Total, api.Stats.data.Total)
	// fmt.Printf("statscopy %T %p %v\n", stat.Total, &stat.Total, stat.Total)
	api.Stats.mu.RUnlock()
	return *stat
}

func (api *apiService) StatsUpdate(s stats) {
	// 	debug("su %v", *api)
	api.Stats.mu.Lock()
	defer api.Stats.mu.Unlock()

	api.Stats.data.LastUpdate = time.Now()
	api.Stats.data = s

	return
}

func (api *apiService) Up(b bool) {
	api.Stats.mu.Lock()
	api.Stats.up = b
	api.Stats.mu.Unlock()
}

func (api *apiService) Status() bool {
	api.Stats.mu.Lock()
	defer api.Stats.mu.Unlock()
	return api.Stats.up

}

// Monitor Starts monitoring Stak
func (api *apiService) Monitor(m *metrics) bool {
	errChan := make(chan error, 10)
	go func(err chan error, api *apiService) {
		timeout := time.Duration(100 * time.Millisecond)
		client := http.Client{
			Timeout: timeout,
		}
		timeoutError := "request canceled"

		for api.Signal != nil {
			select {
			case _, ok := <-api.Signal:
				if !ok {
					return
				}
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
					api.Up(false)
					debug("monitor() %v", err)
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
				api.Up(true)
				hostname, herr := os.Hostname()
				if herr != nil {
					log.Errorf("failed to set hostname %v", herr)
				}

				// general stats
				if met := out.Map(); err == nil {
					// cleanup output, drop port
					poolElement := met["Pool"].(string)
					if strings.Contains(poolElement, ":") {
						met["Pool"] = strings.Split(poolElement, ":")[0]
					}

					tags := map[string]string{"server": hostname}
					err = m.Write("metrics", tags, met)
					if err != nil {
						debug("failed to write to influx %v", err)
					}
				}

				// hashrate tagged by thread id
				if met := out.threadMapSlice(); len(met) != 0 {
					for _, v := range met {
						tags := map[string]string{"server": hostname}
						if v["thread"] != nil {
							tags["thread"] = fmt.Sprintf("%v", v["thread"])
							delete(v, "thread")
						}
						err = m.Write("metrics", tags, v)
						if err != nil {
							debug("failed to write to influx %v", err)
						}
					}
				}

			case outerr := <-errChan:
				debug("%v", outerr)
			}
		}
	}(errChan, api)

	return true
}

func (api *apiService) stopMonitor(m *metrics) bool {
	debug("stopping Stats Service")
	api.Signal <- true
	m.Stop()
	return true
}

func (api *apiService) showMonitor() {
	limit := newLimiter(1005 * time.Millisecond)
	go limitClock(limit)

	for {
		select {
		case <-limit.throttle:
			a := api.StatsCopy()
			if api.Status() {
				a.ConsoleDisplay()
			}
		case _, ok := <-api.Signal:
			if !ok {
				return
			}
			limit.Stop()
			return
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
	up   bool
}

// Map returns a map version of stats data for metrics.go
// non concurrent usage
func (stats *stats) Map() map[string]interface{} {
	if len(stats.Total) == 0 {
		stats.Total = []float64{0}
	}
	m := map[string]interface{}{
		"DiffCurrent": stats.DiffCurrent,
		"SharesGood":  stats.SharesGood,
		"SharesTotal": stats.SharesTotal,
		"AvgTime":     stats.AvgTime,
		"HashesTotal": stats.HashesTotal,
		"Pool":        stats.Pool,
		"Uptime":      stats.Uptime,
		"Ping":        stats.Ping,
		"TotalHR":     stats.Total[0],
	}

	for k, v := range stats.Threads {
		m[fmt.Sprintf("Thread.%v", k)] = v[0]
	}
	return m
}

// Map returns a map version of stats data for metrics.go
// non concurrent usage
func (stats *stats) threadMapSlice() []map[string]interface{} {
	ms := make([]map[string]interface{}, 0, 10)
	for k, v := range stats.Threads {
		ms = append(ms, map[string]interface{}{"hashrate": v[0], "thread": k})

	}
	return ms
}

func (stats *stats) ConsoleDisplay() {
	// 	debug("%T %p %v", stats, stats, stats)
	tm.Clear()
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
	threads := tm.NewTable(0, 10, 5, ' ', 0)

	// headers
	_, _ = fmt.Fprintf(threads, "Thread\tHashrate\n")

	// add threads
	for i, v := range stats.Threads {
		_, _ = fmt.Fprintf(threads, "%v\t%v\n", i, v[0])
	}

	// setup results table
	_, _ = tm.Println(tm.Color("Results", tm.YELLOW))
	ds := tm.NewTable(0, 5, 2, ' ', 0)
	// _, _ = fmt.Fprintf(ds,	"Starting Hash Rate","$script:maxhash H/s")
	// _, _ = fmt.Fprintf(ds,"Restart Hash Rate"="$script:rTarget H/s"

	_, _ = fmt.Fprintln(ds, "Total H/R\t", stats.Total[0])
	// _, _ = fmt.Fprintf(ds,"Minimum Hash Rate"="$script:minhashrate H/s"
	// _, _ = fmt.Fprintf(ds,"Monitoring Uptime"="$tmRunTime"

	_, _ = fmt.Fprintln(ds, "Pool\t", stats.connection.Pool)
	_, _ = fmt.Fprintln(ds, "Uptime\t", stats.connection.Uptime)
	_, _ = fmt.Fprintln(ds, "Difficulty\t", stats.DiffCurrent)
	_, _ = fmt.Fprintln(ds, "Total Shares\t", stats.SharesTotal)
	_, _ = fmt.Fprintln(ds, "Good Shares\t", stats.SharesGood)
	// _, _ = fmt.Fprintf(ds,"Good Share Percent"
	_, _ = fmt.Fprintln(ds, "Share Time\t", stats.AvgTime)
	_, _ = fmt.Fprintln(ds)

	_, _ = tm.Println(ds)
	_, _ = tm.Println(threads)
	TmFlush() // Call it every time at the end of rendering
	// fmt.Printf("%+v", stats)
}

/*
Function refresh-Screen{
			$tmRunTime=get-RunTime -sec ($runTime )
			$tpUpTime=get-RunTime -sec ($script:UpTime )


			if( ( $script:validSensorTime-eq'True' )-and($script:lastRoomTemp ) ){
				$script:displayOutput2+=@{"Last Temp Reading"=@{"$( $script:lastRoomTemp.Time )"="$( $script:lastRoomTemp.$TEMPerSensorLocation ) C"}.ToDisplayString()}
			}

			if( $script:coins ){
				$script:displayOutput2+=show-Coin-Info
			}

			if( $profitLiveCheckingEnabled-eq'True' ){
				$now=(get-date )

				$nextCheck=($script:profitCheckDateTime ).AddMinutes( $ProfitCheckMinutes )
				$countdown=[math]::Round( ($nextCheck-$now ).TotalSeconds, 0 )
				$tFormat=get-RunTime -sec ($countdown )
				$script:displayOutput2+=@{"Last Profit check"=($script:profitCheckDateTime )}
				$script:displayOutput2+=@{"Next Profit check"=$nextCheck}
				$script:displayOutput2+=@{"Time Now"=$now}
				$script:displayOutput2+=@{"Next Profit check due in "="$tFormat"}
			}

			Clear-Host
			Write-Host -fore Green $script:displayOutput2.ToDisplayString()
			if( $slackPeriodicReporting ){
				display-to-slack
			}
		}

*/

func TmFlush() {
	safeFlush.Lock()
	tm.Flush()
	safeFlush.Unlock()
}

func simApi(api *apiService, wg *sync.WaitGroup, startHashRate int, decayRate float64, decayTime time.Duration) {
	ticker := time.NewTicker(decayTime)
	defer ticker.Stop()
	timeout := time.Now().Add(time.Second * 30)

	stat := rwStats{}
	stat.data.Total = []float64{float64(startHashRate)}

	api.StatsUpdate(stat.data)
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
			// 		fmt.Printf("simapi     %T %p %v\n", stat.Total[0], &stat.Total[0], stat.Total[0])
			// 		fmt.Printf("stp        %T %p %v\n", st.Total, &st.Total, st.Total)

			if len(st.Total) == 0 {
				st.Total = []float64{0.0}
			}
			stat.mu.Lock()
			stat.data.Total[0] = st.Total[0] / decayRate
			stat.mu.Unlock()
			api.StatsUpdate(stat.data)
			if stat.data.Total[0] <= 10 {
				return
			}
		}
	}

}
