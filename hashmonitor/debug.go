// +build debug

package hashmonitor

import (
	"fmt"
)

// Debug global debugging variable
var Debug bool

func debug(f string, args ...interface{}) {
	fmt.Println("DEBUG")
	fmt.Printf(f, args...)

}

func init() {
	Debug = true
}
