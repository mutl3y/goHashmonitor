// +build debug

package hashmonitor

import (
	"fmt"
)

func debug(f string, args ...interface{}) {
	fmt.Printf(f, args...)

}
