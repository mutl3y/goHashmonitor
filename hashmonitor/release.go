// +build !debuglocaL

package hashmonitor

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"sync"
	"time"
)

type DebugGrouper struct {
	mu            *sync.RWMutex
	debugMessages chan string
	groupByTime   time.Duration
	stop          chan bool
}

// Debug global debugging variable
var Debug bool

var hostname string

var debugMessages = &DebugGrouper{}

func init() {
	err := error(nil)
	hostname, err = os.Hostname()
	if err != nil {
		hostname = "noHost"
	}
	debugMessages.mu = &sync.RWMutex{}
}

func NewDebugGrouper(v *viper.Viper) *DebugGrouper {
	// d := &DebugGrouper{}
	debugMessages.configDebugGrouper(v)
	go debugMessages.backGroundWriter()

	return debugMessages
}

func debug(f string, args ...interface{}) {
	line := fmt.Sprintf(f+"\n", args...)
	debugMessages.write(line)
}

func (dm *DebugGrouper) write(str string) {
	dm.mu.Lock()
	if dm.debugMessages != nil {
		select {
		case dm.debugMessages <- str:
		case <-time.After(time.Second):
		}
	}
	dm.mu.Unlock()
}

func (dm *DebugGrouper) configDebugGrouper(v *viper.Viper) {
	dm.mu.Lock()
	dm.stop = make(chan bool)
	dm.debugMessages = make(chan string, 1000)
	dm.groupByTime = v.GetDuration("Core.DebugMessageInterval")
	log.Debugf("consolidating debug messages, will flush every %v", dm.groupByTime)
	dm.mu.Unlock()
}

func (dm *DebugGrouper) backGroundWriter() {

	var nextFlush time.Time
	type queue struct {
		points []string
		sync.Mutex
	}
	q := new(queue)
	blankPoints := make([]string, 0, 1000)
	q.points = blankPoints

	flush := func(q *queue) {
		nextFlush = time.Now().Add(dm.groupByTime)
		q.Lock()
		length := len(q.points)
		if length > 0 {
			p := q.points
			q.points = make([]string, 0, 1000)

			go func(p []string) {
				currentTime := time.Now().Format("2006-01-02 15:04:05 MST")
				s := fmt.Sprintf("%v: %v \n", hostname, currentTime)
				for _, v := range p {
					if s != "" {
						s += "\x1b[0;35m" + v
					}
				}
				log.Debug(s)
			}(p)

		}

		q.Unlock()
	}

	for {
		select {
		case p, ok := <-dm.debugMessages:
			if !ok {
				flush(q)
				fmt.Printf("flush case fail debug msg queue %v", len(q.points))
				return
			}
			q.Lock()
			q.points = append(q.points, p)
			q.Unlock()
		case <-afterTime(nextFlush):
			flush(q)

		case <-dm.stop:

			return
		}
	}
}
func (dm *DebugGrouper) Stop() {
	dm.stop <- true
}
