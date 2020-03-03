package zhanio

import (
	"math/rand"
	"testing"
	"time"
)

func TestBytePool(t *testing.T) {
	n := 5
	per := 3
	bytePool := InitBytePool(n)
	cur := -1
	for i := 0; i < per*n; i++ {
		if i%per == 0 {
			cur = cur + 1
		}
		go func(index int, curNumber int) {
			for {
				bytes := bytePool.Get(curNumber*1024 + rand.Intn(1024))
				println(index, curNumber, bytes)
				time.Sleep(time.Millisecond * 500)
				bytePool.Put(bytes)
			}
		}(i, cur)
	}
	time.Sleep(100 * time.Second)
}
