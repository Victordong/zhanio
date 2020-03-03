package zhanio

import "sync"

type BufferPool struct {
	sync.Pool
}

func InitBufferPool() *BufferPool {
	var bufferPool BufferPool
	bufferPool.New = func() interface{} {
		return &RingBuffer{
			wPos:    0,
			rPos:    0,
			isEmpty: false,
		}
	}
	return &bufferPool
}
