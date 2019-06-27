package hashmonitor

import (
	"fmt"
	"os"

	"github.com/gogap/logrus_mate"
	_ "github.com/gogap/logrus_mate/hooks/expander"
	_ "github.com/gogap/logrus_mate/hooks/file"
	_ "github.com/gogap/logrus_mate/hooks/slack"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// default logger goes to std.out
var log = logrus.StandardLogger()

func ConfigLogger(fn string, force bool) (err error) {
	c := "hashmonitor"
	_, err = os.Stat(fn)
	if os.IsNotExist(err) || force {
		err = defaultLoggerConfig(fn)
		if err != nil {
			return errors.Wrapf(err, "%v", fn)
		}
	}

	// recover from 3rd party code panics
	defer func() {
		if r := recover(); r != nil {
			e := fmt.Sprintf("invalid config file %v", r)
			err = errors.New(e)
			fmt.Println("recovered from logrus_mate config issue")
		}
	}()
	cfn := logrus_mate.ConfigFile(fn)

	mate, err := logrus_mate.NewLogrusMate(cfn)
	if err != nil {
		fmt.Println("log config error")
		return errors.Wrapf(err, "%v", mate)
	}

	if err = mate.Hijack(log, c); err != nil {
		return errors.Wrapf(err, "failed to find log config '%v' in %v\n", c, fn)

	}

	log.SetReportCaller(false)
	return errors.Wrapf(err, "%v", mate)
}

func defaultLoggerConfig(fn string) error {
	f, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Wrapf(err, "defaultLoggerConfig %v", fn)
	}
	_, err = f.WriteString(conf)
	if err != nil {
		return errors.Wrap(err, "defaultLoggerConfig failed to write default config")
	}
	return errors.Wrapf(f.Close(), "defaultLoggerConfig failed to close file %v", err)
}

var conf = `
hashmonitor{
	level = "debug"
	formatter.name = "text"
	formatter.options{
		force-colors = true
		disable-colors = false
		disable-timestamp = false
		full-timestamp = false
		timestamp-format = "2006-01-02 15:04:05"
		disable-sorting = false
		}

hooks{
	expander{}
	file{
		level = 3
		filename = "hashmonitor.log"
		daily = true
		rotate = false
		maxlines = 10000
		maxdays = 2
		perm = 0600
		maxsize  = 1024
		}
	slack {
        url      = "https://hooks.slack.com/services/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"
        levels   = ["error", "info", "warn"]
        channel  = ""
        emoji    = ":rag:"
        username = "logrus_mate"
        }
	}
}
`

// todo replace slack url with std goHashmonitor version
