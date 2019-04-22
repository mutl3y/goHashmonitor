// +build debug

package hashmonitor

import (
	"fmt"
)

// Debug global debugging variable
var Debug bool

func debug(f string, args ...interface{}) {
	fmt.Printf("DEBUG: "+f, args...)

}

func init() {
	Debug = true
}
