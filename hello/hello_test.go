package main

import (
	"testing"
	"time"
)

type comms struct {
	ch []chan int
}

func (c *comms) worker(n int) {
	for i := range n {
		select {
		case <-c.ch[0]:
			//println("recv 0")
		case <-c.ch[1]:
			//println("recv 1")
		case <-c.ch[2]:
			//println("recv 2")

		case c.ch[0] <- i:
			//println("send 0")
		case c.ch[1] <- i:
			//println("send 1")
		case c.ch[2] <- i:
			//println("send 2")
		}
	}
}

func Test001(t *testing.T) {
	c := &comms{}
	n := 3
	for range n {
		c.ch = append(c.ch, make(chan int))
	}

	for range n {
		go c.worker(100_000_000)
	}

	time.Sleep(time.Second)
	println("goodbye")
}
