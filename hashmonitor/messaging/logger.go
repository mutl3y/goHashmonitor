package messaging

import (
	"github.com/spf13/viper"
	"log"
	"os"
	"strings"
	"time"
)

var pathSep = string(os.PathSeparator)

type logger struct {
	msg        string
	mtype      string
	attachment string
	color      string
	silent     bool
}

type Loogger interface {
	ToFile()
}

func NewLog(c *viper.Viper) *logger {
	return &logger{}
}

func toFile(c *viper.Viper) {

	ld := c.GetString("Core.Log.Dir")
	lf := c.GetString("Core.Log.File")
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	dir := strings.Join([]string{cwd, ld}, pathSep)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 666)
		if err != nil {
			log.Fatalf("failed to create directory %v \t%v", dir, err)
		}

	}

	file := strings.Join([]string{dir, time.Now().Format("2006-01-02_") + lf}, pathSep)
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}
	line := time.Now().Format("2006-01-02_15:04:05,") + "\n"
	if _, err := f.Write([]byte(line)); err != nil {
		log.Fatal(err)
	}

}

func dispatcher() {}
