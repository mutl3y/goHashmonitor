package messaging

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var pathSep = string(os.PathSeparator)

type msg struct {
	text       string
	queue      string
	mtype      string
	attachment string
	priority   int
	silent     bool
}

// func (m msg) String() string {
// 	return fmt.Sprintf("text: \t\t%v\nqueue: \t\t%v\ntype: \t\t%v\nattachment: %v\npriority: \t%v\nsilent: \t%v", m.text, m.queue, m.mtype, m.attachment, m.priority, m.silent)
// }

type config struct {
	logDir, logName              string
	slack, telegram, file, email bool
}

type logger struct {
	msgChan chan msg
	control chan string
	config  config
}

// NewLogger  returns a new logger instance
func NewLogger(c *viper.Viper) *logger {
	conf := config{}
	conf.logDir = c.GetString("Core.Log.Dir")
	conf.logName = c.GetString("Core.Log.File")
	l := &logger{}
	l.control = make(chan string, 1)
	l.msgChan = make(chan msg, 10)
	return l
}

// Send a message
func (l *logger) Send(m msg) error {
	fmt.Printf("sending  message %v\n", m)
	l.msgChan <- m
	return nil
}

// func sendFile(m msg) {
//
// }
//
// func NewLog(c *viper.Viper) *logger {
// 	return &logger{}
// }

func (l *logger) fileWriter(chan msg) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	dir := strings.Join([]string{cwd, l.config.logDir}, pathSep)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 666)
		if err != nil {
			log.Fatalf("failed to create directory %v \t%v", dir, err)
		}

	}
	f := &os.File{}

	for {
		select {
		case m, ok := <-l.msgChan:
			if !ok {
				return
			}
			file := strings.Join([]string{dir, time.Now().Format("2006-01-02_") + l.config.logName}, pathSep)
			if file != f.Name() {
				// file.close
				if err := f.Close(); err != nil {
					fmt.Printf("failed closing file %v %v", f.Name(), err)
				}
			}

			// If the file doesn't exist, create it, or append to the file
			f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

			if err != nil {
				log.Fatal(err)
			}
			line := time.Now().Format("2006-01-02_15:04:05,") + m.text
			if _, err := f.Write([]byte(line)); err != nil {
				log.Fatal(err)
			}
		}
		// /	defer f.Close()
	}
}

func (l *logger) dispatcher() {
	fmt.Printf("Starting dispatcher\n")
	l.config.file = true
	for {
		select {
		case m, ok := <-l.msgChan:
			if !ok {
				fmt.Printf("no more messages to dispatch\n")
				return
			}
			fmt.Printf("recieved message %v\n", m)
			if l.config.slack {
				fmt.Printf("sending to slack\n")
				// 			toSlack(msg)
			}
			if l.config.file {
				fmt.Printf("sending to file\n")
				// 			toFile(msg)
			}
			if l.config.telegram {
				fmt.Printf("sending to telegram\n")
				// 			toTelegram(msg)
			}
			if l.config.email {
				fmt.Printf("sending to email\n")
				// 			toEmail()
			}
		case <-l.control:
			fmt.Printf("Recieved stop signal, exiting in 5 seconds\n")
			// l.msgChan = nil
			time.Sleep(5 * time.Second)
			return
		}
	}
}
