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
	StopMonitor(met *metrics) bool
	ShowMonitor()
}
type apiService struct {
	URL    string
	Signal chan bool
	limit  *simpleRateLimit
	Stats  *rwStats
	hrMon  hrMon
}

// newStatsService returns a monitoring service with rate limiter
// takes settings from viper config
func NewStatsService(cfg *viper.Viper) ApiService {
	Signal := make(chan bool)
	apiIp := cfg.GetString("Core.Stak.Ip")
	apiPort := cfg.GetInt64("Core.Stak.Port")
	apiUrl := fmt.Sprintf("http://%v:%v/Api.json", apiIp, apiPort)

	debug("newStatsService, %v:%v ", apiIp, apiPort)
	hrmon := NewhRMonStruct()
	hrmon.drop = cfg.GetInt("Core.Hash.Drop")

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
		hrMon:  hrmon,
	}
}

func (api *apiService) StatsCopy() stats {

	api.Stats.mu.RLock()
	defer api.Stats.mu.RUnlock()
	stat := stats{}
	stat = api.Stats.data
	if len(stat.Total) == 0 {
		stat.Total = []float64{0}
	}
	// fmt.Printf("Api %T %p %v\n", Api.Stats.data.Total, &Api.Stats.data.Total, Api.Stats.data.Total)
	// fmt.Printf("statscopy %T %p %v\n", stat.Total, &stat.Total, stat.Total)

	return stat
}

func (api *apiService) StatsUpdate(s stats) {

	api.Stats.mu.Lock()
	api.Stats.data.LastUpdate = time.Now()
	api.Stats.data = s
	api.Stats.mu.Unlock()
	return
}

func (api *apiService) Up(b bool) {
	api.Stats.mu.Lock()
	api.Stats.up = b
	api.Stats.statusChangeTime = time.Now()
	api.Stats.mu.Unlock()
}

func (api *apiService) Status() (ok bool) {
	api.Stats.mu.RLock()
	defer api.Stats.mu.RUnlock()
	ok = api.Stats.up
	return
}

func (api *apiService) CheckApi(checks int, sleepTime time.Duration) error {

	fmt.Printf("checking api ")
	for i := 0; i <= checks; i++ {
		d := api.Status()
		if d {
			return nil
		}
		fmt.Printf(".")
		time.Sleep(sleepTime)
	}
	fmt.Println()
	return fmt.Errorf("stak has stopped responding")
}

func (api *apiService) Monitor(m *metrics) bool {
	errChan := make(chan error, 10)
	go func(errChan chan error, api *apiService) {
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
				errChan <- fmt.Errorf("Service.Monitor signal channel closed")
				close(errChan)
				return
			case <-api.limit.throttle:
				out := stats{}
				res, err := client.Get(api.URL)
				if err != nil {
					if strings.Contains(err.Error(), timeoutError) {
						api.StatsUpdate(out)
						api.Up(false)
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

func (api *apiService) StopMonitor(m *metrics) bool {
	debug("stopping Stats Service")
	api.Signal <- true
	m.Stop()
	return true
}

func (api *apiService) ShowMonitor() {
	limit := newLimiter(1005 * time.Millisecond)
	go limitClock(limit)

	for {
		select {
		case <-limit.throttle:
			api.ConsoleDisplay()
		case _, ok := <-api.Signal:
			if !ok {
				return
			}
			limit.Stop()
			return
		}
	}
}

// stak Api data structs
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
	mu               sync.RWMutex
	data             stats
	up               bool
	statusChangeTime time.Time
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

func (api *apiService) ConsoleDisplay() {
	api.hrMon.mu.RLock()
	defer api.hrMon.mu.RUnlock()
	if api.hrMon.startingHash <= 1 {
		return
	}

	// 	debug("%T %p %v", stats, stats, stats)
	api.Stats.mu.RLock()
	defer api.Stats.mu.RUnlock()
	s := api.Stats.data

	tm.Clear()
	tm.MoveCursor(1, 1)
	ct := fmt.Sprintf("Current Time: %v", time.Now().Format(time.RFC1123))
	_, _ = tm.Println(tm.Color(ct, tm.GREEN))
	_, _ = tm.Println()

	// normalise fields for printing
	if len(s.Total) == 0 {
		s.Total = []float64{0}
	}

	if s.connection.Pool == "" {
		s.connection.Pool = "Not Connected"
	}

	// setup Threads table
	threads := tm.NewTable(0, 10, 5, ' ', 0)

	// headers
	_, _ = fmt.Fprintf(threads, "Thread\tHashrate\n")

	// add threads
	for i, v := range s.Threads {
		_, _ = fmt.Fprintf(threads, "%v\t%v\n", i, v[0])
	}

	// setup results table
	_, _ = tm.Println(tm.Color("Results", tm.YELLOW))
	ds := tm.NewTable(0, 5, 2, ' ', 0)
	// _, _ = fmt.Fprintf(ds,	"Starting Hash Rate","$script:maxhash H/s")
	// _, _ = fmt.Fprintf(ds,"Restart Hash Rate"="$script:rTarget H/s"

	_, _ = fmt.Fprintln(ds, "Total H/R\t", s.Total[0])

	_, _ = fmt.Fprintln(ds, "Minimum Hash Rate\t", api.hrMon.min())

	// _, _ = fmt.Fprintf(ds,"Monitoring Uptime"="$tmRunTime"

	_, _ = fmt.Fprintln(ds, "Pool\t", s.connection.Pool)
	_, _ = fmt.Fprintln(ds, "Uptime\t", time.Duration(float64(s.connection.Uptime))*time.Second)
	_, _ = fmt.Fprintln(ds, "Difficulty\t", s.DiffCurrent)
	_, _ = fmt.Fprintln(ds, "Total Shares\t", s.SharesTotal)
	_, _ = fmt.Fprintln(ds, "Good Shares\t", s.SharesGood)
	// _, _ = fmt.Fprintf(ds,"Good Share Percent"
	_, _ = fmt.Fprintln(ds, "Share Time\t", s.AvgTime)
	_, _ = fmt.Fprintln(ds)

	_, _ = tm.Println(ds)
	_, _ = tm.Println(threads)
	tmFlush() // Call it every time at the end of rendering
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

var sem = make(chan bool, 1)

func tmFlush() {
	sem <- true
	tm.Flush()
	// time.Sleep(50 * time.Millisecond) //todo do io need this
	<-sem
}

func simApi(api *apiService, wg *sync.WaitGroup, startHashRate int, decayRate float64, decayTime time.Duration) {
	ticker := time.NewTicker(decayTime)
	defer ticker.Stop()
	timeout := time.Now().Add(time.Second * 30)

	stat := &rwStats{}
	stat.data.Total = []float64{float64(startHashRate)}

	wg.Done()
	for {
		select {
		case <-afterTime(timeout):
			return
		case <-ticker.C:

			st := api.StatsCopy()
			stat.mu.Lock()
			stat.data.Total = []float64{st.Total[0] / decayRate}
			api.StatsUpdate(stat.data)
			if stat.data.Total[0] <= 10 {
				stat.mu.Unlock()
				return
			}
			stat.mu.Unlock()

		}
	}

}
