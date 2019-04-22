package hashmonitor

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"net/url"
	"sync"
	"time"

	inf "github.com/influxdata/influxdb1-client"
)

type metrics struct {
	client      *inf.Client
	db          string
	refresh     time.Duration
	pointsQueue chan inf.Point
	done        chan bool
	enabled     bool
}

type Metrics interface {
	Ping() (time.Duration, error)
	Write(measurment string, tags map[string]string, fields map[string]interface{}) error
	Query(map[string]interface{}) error
}

func NewMetricsClient() *metrics {
	m := new(metrics)
	m.pointsQueue = make(chan inf.Point, 300)
	m.done = make(chan bool)
	return m
}

// Config Configure metrics client using config from viper
// Viper keys
// "Influx.Enabled" bool
// "Influx.Port" int64
// "Influx.Ip" string
// "Influx.User" string
// "Influx.Pw" string
// "Influx.DB" string
// "Influx.FlushSec" time.duration
func (m *metrics) Config(c *viper.Viper) error {
	if !c.GetBool("Influx.Enabled") {
		return nil
	}
	m.enabled = true

	ip := c.GetString("Influx.Ip")
	if ip == "" {
		ip = "127.0.0.1"
	}

	port := c.GetInt64("Influx.Port")
	if port == 0 {
		port = 8086
	}

	host, err := url.Parse(fmt.Sprintf("http://%s:%d", ip, port))
	if err != nil {
		return fmt.Errorf("failed to parse Influx url %v", err)
	}

	user := c.GetString("Influx.User")
	pw := c.GetString("Influx.Pw")

	conf := inf.Config{
		URL:      *host,
		Username: user,
		Password: pw,
	}

	m.client, err = inf.NewClient(conf)
	if err != nil {
		return errors.Wrap(err, "NewClient")
	}

	m.refresh = c.GetDuration("Influx.FlushSec")
	if m.refresh < time.Second*10 {
		m.refresh = time.Second * 10
	}

	m.db = c.GetString("Influx.DB")
	if m.db == "" {
		log.Infof("failed to set Influx DB")
		m.db = "hashmonitor"
	}

	return nil
}

func (m *metrics) Ping() (time.Duration, error) {
	dur, _, err := m.client.Ping()
	if err != nil {
		return 0, err
	}
	return dur, nil
}

func (m *metrics) Query(c, d string) (string, error) {
	q := inf.Query{
		Command:  c,
		Database: d,
	}
	response, err := m.client.Query(q)
	resString := fmt.Sprintf("%v", response.Results[0])
	if err != nil && response.Error() != nil {
		log.Println(response.Results)
		return "", err
	}

	return resString, nil
}

func (m *metrics) Write(measurment string, tags map[string]string, fields map[string]interface{}) error {
	if !m.enabled {
		return nil
	}
	p := inf.Point{
		Measurement: measurment,
		Tags:        tags,
		Fields:      fields,
		Time:        time.Now(),
	}

	select {
	case <-time.After(time.Millisecond * 100):
		log.Infof("influx write queue timed out")
		time.Sleep(time.Millisecond * 1000)

		return errors.New("stats queue full, discarding")
	case m.pointsQueue <- p:
	}
	return nil
}

func (m *metrics) checkDB() error {
	// turn call into a no op if not enabled
	if !m.enabled {
		log.Infof("metrics disabled")
		return nil
	}
	query := inf.Query{
		Command:  fmt.Sprintf("CREATE DATABASE %s", m.db),
		Database: m.db,
	}
	results, err := m.client.Query(query)
	if err != nil {
		log.Fatalf("error using Influx, %v", err)

	}

	if results.Err != nil {
		return fmt.Errorf("failed creating DB")

	}

	debug("checkDB results %v", results)
	return err
}

func (m *metrics) afterTime(t *time.Time) <-chan time.Time {
	var C chan time.Time
	if time.Now().After(*t) {
		*t = time.Now().Add(time.Second * 2)
		return time.After(time.Duration(0))
	}

	return C
}

var nextFlush time.Time

func (m *metrics) backGroundWriter() {
	// turn call into a no op if not enabled
	if !m.enabled {
		return
	}

	log.Info("Starting background Influx writer")
	type queue struct {
		points []inf.Point
		sync.RWMutex
	}
	q := new(queue)
	blankPoints := make([]inf.Point, 0, 100)
	q.points = blankPoints

	flush := func(q *queue) {
		length := len(q.points)
		if length >= 1 {
			q.Lock()
			p := q.points
			q.points = blankPoints
			q.Unlock()
			go func(p []inf.Point) {
				res, err := m.client.Write(inf.BatchPoints{Points: p, Database: m.db, RetentionPolicy: "autogen", Time: time.Now()})
				if err != nil {
					log.Errorf("backGroundWriter: %v", err)
				}
				debug("client write: %+v\n", p)
				debug("response %v\n", res)
			}(p)
		}
	}

	for {
		select {
		case p, ok := <-m.pointsQueue:
			if !ok {
				log.Debugf("backGroundWriter: influx queue closed")
				flush(q)
				log.Debugf("Stopping background influx writer")
				time.Sleep(time.Second * 2)
				return
			}
			q.Lock()
			q.points = append(q.points, p)
			q.Unlock()
		case <-m.afterTime(&nextFlush):
			flush(q)

		case _, ok := <-m.done:
			if !ok {
				log.Debugf("backGroundWriter: done called")
				flush(q)
				time.Sleep(time.Second * 2)
				return
			}

		}
	}
}

//  stats to send
/*  stats to send
threads := []int{122, 3321, 4434, 5655, 666, 777}


for i, v := range threads {
	id := fmt.Sprintf("thread_%v", i)
	pts[0].Fields[id] = v
	stats.Set(id, v)
}

res, err	:= m.client.Write(bps)
if err != nil {
	log.Fatal(err)
}

for k, v := range stats.Items() {
	fmt.Printf("%v:%v\n", k, v)
}
return res.Err-

Function grafana{
$Metrics.add( 'balance', $script:balance )
$Metrics.add( 'btcprice', $script:btcprice )
$Metrics.add( "estCoin$coinStats", $script:coins )
$Metrics.add( "estDollar$coinStats", $script:dollars )
$Metrics.add( 'avghash1hr', $script:avghash1hr )
$script:nanopoolLastUpdate=$runTime
}*/