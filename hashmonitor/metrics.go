package hashmonitor

import (
	"fmt"
	"github.com/pkg/errors"
	"net/url"
	"os"
	"sync"
	"time"

	inf "github.com/influxdata/influxdb1-client"
	"github.com/orcaman/concurrent-map"
)

var (
	stats      = cmap.New()
	influxUrl  = fmt.Sprintf("http://%s:%d", "192.168.0.29", 8086)
	influxUdp  = 8089
	InfluxHttp = 8086
	infClient  inf.Client
)

func SendStats(m map[string]interface{}) {

}

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
	Write(map[string]interface{}) error
	Query(map[string]interface{}) error
}

//type Influx interface {
//	Ping()
//	Write()
//	Query()
//}

//func NewClient() (*inf.Client, error) {
//	host, err := url.Parse("")
//	if err != nil {
//		return &inf.Client{}, err
//	}
//
//	// NOTE: this assumes you've setup a user and have setup shell env variables,
//	// namely INFLUX_USER/INFLUX_PWD. If not just omit Username/Password below.
//	conf := inf.Config{
//		URL:      *host,
//		Username: os.Getenv("INFLUX_USER"),
//		Password: os.Getenv("INFLUX_PWD"),
//	}
//
//	c, err := inf.NewClient(conf)
//	if err != nil {
//		return nil, errors.Wrap(err,"NewClient")
//	}
//	return c, nil
//}

func NewMetricsClient() *metrics {
	m := new(metrics)
	m.pointsQueue = make(chan inf.Point, 300)
	m.done = make(chan bool)
	return m
}

func (m *metrics) Config(u string) error {
	if !cfg.GetBool("Influx.Enabled") {
		return nil
	}
	m.enabled = true
	host, err := url.Parse(u)
	if err != nil {
		return err
	}
	// NOTE: this assumes you've setup a user and have setup shell env variables,
	// namely INFLUX_USER/INFLUX_PWD. If not just omit Username/Password below.
	conf := inf.Config{
		URL:      *host,
		Username: os.Getenv("INFLUX_USER"),
		Password: os.Getenv("INFLUX_PWD"),
	}

	m.client, err = inf.NewClient(conf)
	if err != nil {
		return errors.Wrap(err, "NewClient")
	}
	m.db = cfg.GetString("Influx.DB")
	m.refresh = cfg.GetDuration("Influx.FlushSec")
	if m.refresh < time.Second*10 {
		m.refresh = time.Second * 10
	}

	m.db = cfg.GetString("Influx.DB")
	if m.db == "" {
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
		log.Debug("influx write queue timed out")
		time.Sleep(time.Millisecond * 1000)

		return errors.New("stats queue full, discarding")
	case m.pointsQueue <- p:
		//fmt.Printf("%+v", p)
	}
	return nil
}

func (m *metrics) checkDB() error {
	// turn call into a no op if not enabled
	if !m.enabled {
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

	return err
}

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
	q := &queue{}
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
					log.Errorf("failed writing to influx %v\n", err)
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
				log.Debugf("influx queue closed")
				flush(q)
				log.Debugf("Stopping background influx writer")
				time.Sleep(time.Second * 2)
				return
			}
			q.Lock()
			q.points = append(q.points, p)
			q.Unlock()
		case <-time.After(time.Second * m.refresh):
			flush(q)
		case _, ok := <-m.done:
			if !ok {
				log.Debugf("background influx writer close called")
				flush(q)
				time.Sleep(time.Second * 2)
				return
			}

		}
	}
}

//threads := []int{122, 3321, 4434, 5655, 666, 777}
//
//for i, v := range threads {
//	id := fmt.Sprintf("thread_%v", i)
//	pts[0].Fields[id] = v
//	stats.Set(id, v)
//}

//
//res, err	:= m.client.Write(bps)
//if err != nil {
//	log.Fatal(err)
//}
//for k, v := range stats.Items() {
//	fmt.Printf("%v:%v\n", k, v)
//}
//return res.Err-

/*Function grafana{

$Metrics.add( 'balance', $script:balance )
$Metrics.add( 'btcprice', $script:btcprice )
$Metrics.add( "estCoin$coinStats", $script:coins )
$Metrics.add( "estDollar$coinStats", $script:dollars )
$Metrics.add( 'avghash1hr', $script:avghash1hr )
$script:nanopoolLastUpdate=$runTime

}*/
