package main

import (
	"testing"
	"testing/synctest"
	"time"
)

func sleeper(dur time.Duration) {
	time.Sleep(dur)
	println("sleep done", dur)
}

func Test001(t *testing.T) {

	synctest.Run(func() {

		println("hello")

		go sleeper(time.Second)
		go sleeper(time.Second * 2)

		time.Sleep(time.Second * 3)
		println("goodbye")
	})
}
