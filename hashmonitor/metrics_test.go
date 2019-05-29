package hashmonitor

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	influxClient "github.com/influxdata/influxdb1-client"
	"github.com/spf13/viper"
)

var (
	testCfg = viper.New()
	db      = "hashMonitor"
)

func init() {
	testCfg.Set("Influx.Enabled", true)
	testCfg.Set("Influx.DB", db)
	testCfg.Set("Influx.Retention", time.Second*10)
	testCfg.Set("Influx.Ip", "192.168.0.29")
	testCfg.Set("Influx.Port", 8086)

}

func Test_Inf_Client_Write(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		in, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		} else if have, want := strings.TrimSpace(string(in)), `m0,host=server01 v1=2,v2=2i,v3=2u,v4="foobar",v5=true 0`; have != want {
			t.Errorf("unexpected write protocol: %s != %s", have, want)
		}
		var data influxClient.Response
		w.WriteHeader(http.StatusNoContent)
		_ = json.NewEncoder(w).Encode(data)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	config := influxClient.Config{URL: *u}
	c, err := influxClient.NewClient(config)
	if err != nil {
		t.Fatalf("unexpected error.  expected %v, actual %v", nil, err)
	}

	bp := influxClient.BatchPoints{
		Points: []influxClient.Point{
			{
				Measurement: "m0",
				Tags: map[string]string{
					"host": "server01",
				},
				Time: time.Unix(0, 0).UTC(),
				Fields: map[string]interface{}{
					"v1": float64(2),
					"v2": int64(2),
					"v3": uint64(2),
					"v4": "foobar",
					"v5": true,
				},
			},
		},
	}
	r, err := c.Write(bp)
	if err != nil {
		t.Fatalf("unexpected error.  expected %v, actual %v", nil, err)
	}
	if r != nil {
		t.Fatalf("unexpected response. expected %v, actual %v", nil, r)
	}
}

func Test_NewMetricsClient(t *testing.T) {
	tests := []struct {
		name string
	}{
		{""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewMetricsClient()
			if c == nil {
				t.Fatalf("failed to configure influx client")
			}
		})
	}
}

func TestExampleClient_Ping(t *testing.T) {
	tdb := "dbd"
	testCfg.Set("Influx.DB", tdb)
	c := NewMetricsClient()
	err := c.Config(testCfg)

	if err != nil {
		t.Fatalf("failed to configure influx %v", err)
	}
	tests := []struct {
		name string
	}{
		{"one"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err = c.Ping()
			if err != nil {
				t.Fatalf("failed to ping influx %v", err)
			}

		})
	}
}

func TestExampleClient_Query(t *testing.T) {
	c := NewMetricsClient()
	err := c.Config(testCfg)
	if err != nil {
		t.Fatalf("failed to configure influx %v", err)
	}

	tests := []struct {
		name string
	}{
		{""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, qErr := c.Query("Show Series ", "hashmonitorTest")
			if qErr != nil {
				t.Fatalf("%v", qErr)
			}
			t.Logf("response %v", res)
		})
	}
}

func Test_Write(t *testing.T) {
	tdb := "hashmonitorTest2"
	testCfg.Set("Influx.DB", tdb)
	err := ConfigLogger("logging.amdConf", false)
	c := NewMetricsClient()
	err = c.Config(testCfg)
	if err != nil {
		t.Fatalf("failed to configure influx %v", err)
	}

	go c.backGroundWriter()
	defer c.Stop()
	tests := []struct {
		name string
	}{
		{""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for ix := 0; ix <= 10; ix++ {
				ta := rand.Int63n(10000)
				s1 := rand.NewSource(time.Now().UnixNano() / ta)
				r1 := rand.New(s1)
				tags := map[string]string{"server": "hostnamegoeshere"}
				fields := map[string]interface{}{"floats": r1.ExpFloat64()}
				err = c.Write("goHashmonitor", tags, fields)
				if err != nil {
					t.Fatalf("failed to write to influx %v", err)
				}
			}
			// ti := rand.Int()
			ta := rand.Intn(10000)
			time.Sleep(time.Duration(ta))

		})
	}

	// the write function sends into a message queue, need sleep to allow backGroundWriter to do its work
	time.Sleep(2 * time.Second)

}

func Test_metrics_checkDB(t *testing.T) {
	tdb := "dbd"
	testCfg.Set("Influx.DB", tdb)
	c := NewMetricsClient()
	err := c.Config(testCfg)
	if err != nil {
		t.Fatalf("failed to configure influx %v", err)
	}

	type fields struct {
		db        string
		refresh   time.Duration
		retention time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"", fields{"htest", time.Second * 120, time.Second * 120}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err = c.checkDB(); (err != nil) != tt.wantErr {
				t.Errorf("metrics.checkDB() error = %v, match %v", err, tt.wantErr)
			}
		})
	}
}

func Test_metrics_Event(t *testing.T) {
	tdb := "testMine"
	testCfg.Set("Influx.DB", tdb)
	err := ConfigLogger("logging.amdConf", false)
	if err != nil {
		t.Fatalf("amdConf log issue %v", err)
	}
	m := NewMetricsClient()
	err = m.Config(testCfg)
	if err != nil {
		t.Fatalf("metric config error %v", err)
	}

	tests := []struct {
		name    string
		title   string
		text    string
		tags    string
		wantErr bool
	}{
		{"", "first test event", "this is the text stuff", "tags,go,in,here", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = m.Event(tt.title, tt.text, tt.tags)
			if (err != nil) != tt.wantErr {

				t.Errorf("metrics.Event() error = %v, wantErr %v", err, tt.wantErr)

			}
		})
	}
	m.Stop()
}
