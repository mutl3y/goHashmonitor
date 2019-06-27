// +build debuglocaL

package hashmonitor

// Debug global debugging variable
var Debug bool

func debug(f string, args ...interface{}) {
	fmt.Printf("DEBUG: "+f+"\n", args...)
	// log.Debugf(f, args...)

}

func init() {
	// Debug = true
}
