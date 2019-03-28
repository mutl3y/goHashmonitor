package hashmonitor

import (
	"os"

	"github.com/gogap/logrus_mate"
	_ "github.com/gogap/logrus_mate/hooks/expander"
	_ "github.com/gogap/logrus_mate/hooks/file"
	_ "github.com/gogap/logrus_mate/hooks/slack"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Log application logging
var log = &logrus.Logger{}

func init() {

}

func ConfigLogger(fn string, force bool) error {
	c := "hashmonitor"
	_, err := os.Stat(fn)
	if os.IsNotExist(err) || force {
		err = defaultLoggerConfig(fn)
		if err != nil {
			return errors.Wrapf(err, "%v", fn)
		}
	}

	mate := &logrus_mate.LogrusMate{}
	if mate, err = logrus_mate.NewLogrusMate(logrus_mate.ConfigFile(fn)); err != nil {
		return errors.Wrapf(err, "%v", mate)
	}
	if err = mate.Hijack(log, c); err != nil {
		return errors.Wrapf(err, "failed to find log config '%v' in %v\n", c, fn)

	}
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

var conf = `hashmonitor{
level = "error"
	formatter.name = "text"
	formatter.options{
		force-colors = false
		disable-colors = false
		disable-timestamp = false
		full-timestamp = false
		timestamp-format = "2006-01-02 15:04:05"
		disable-sorting = false
		}

hooks{
	expander{}
	file{
		filename = "hashmonitor.log"
		daily = true
		rotate = true
		}
	}
}
`
