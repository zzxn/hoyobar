package funcs

import (
	"log"
	"runtime/debug"
)

// 捕获panic
func Go(f func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic %v\n", err)
				log.Printf("stack %v\n", debug.Stack())
			}
		}()
		f()
	}()
}
