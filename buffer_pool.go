package zhanio

import "sync"

var BufferPool sync.Pool

func init() {
	BufferPool.New = func() interface{} {
		return &RingBuffer{}
	}
}
