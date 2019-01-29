package main

import (
	"fmt"
	"github.com/sniperHW/cooprative"
	"sync/atomic"
	"time"
)

func main() {

	c1 := int32(0)
	c2 := int32(0)
	count := int32(0)

	tickchan := time.Tick(time.Millisecond * time.Duration(1000))
	go func() {
		for {
			_ = <-tickchan
			tmp := atomic.LoadInt32(&count)
			atomic.StoreInt32(&count, 0)
			fmt.Printf("count:%d\n", tmp)

		}
	}()

	s := cooprative.NewScheduler()

	var fn func(int)

	fn = func(a int) {

		if a != 0 {
			panic("a != 0")
		}

		atomic.AddInt32(&c1, 1)
		atomic.AddInt32(&count, 1)
		c2++
		if c1 != c2 {
			fmt.Printf("not equal,%d,%d\n", c1, c2)
		}

		if c2 >= 1000000 {
			s.Close()
			return
		}

		s.Await(time.Sleep, time.Millisecond*time.Duration(10))

		s.PostFn(fn, 0)
	}

	for i := 0; i < 100000; i++ {
		s.PostFn(fn, 0)
	}

	s.Start()

	fmt.Printf("scheduler stop,total taskCount:%d\n", c2)

}