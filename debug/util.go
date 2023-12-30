package debug

import "log"

const Debug = false

func Dlog(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}
