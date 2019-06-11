// +build !debug

package hashmonitor

// Debug global debugging variable
var Debug bool

func debug(f string, args ...interface{}) {
	log.Debugf(f, args...)
}
