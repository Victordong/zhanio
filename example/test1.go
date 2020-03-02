package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			x := Buffer.Get()
			fmt.Println(x)
			time.Sleep(time.Second)
			Buffer.Put(x)
		}()
	}
	wg.Wait()
}
