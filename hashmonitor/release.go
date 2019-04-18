// +build !debug

package hashmonitor

// Debug global debugging variable
var Debug bool

func debug(fmt string, args ...interface{}) {}
