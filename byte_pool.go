package zhanio

import (
	"sync"
)

type BytePool struct {
	pools     []sync.Pool
	typeNum   int
	sliceSize int
}

const defaultSliceSize = 1024

const defaultTypeNum = 10

func InitBytePool(typeNum int, sliceSize int) *BytePool {
	var bytePool BytePool
	bytePool.pools, bytePool.typeNum, bytePool.sliceSize = make([]sync.Pool, typeNum), typeNum, sliceSize
	for i, _ := range bytePool.pools {
		index := i
		bytePool.pools[i].New = func() interface{} {
			bytes := make([]byte, (index+1)*1024)
			return &bytes
		}
	}
	return &bytePool
}

func (p *BytePool) Get(size int) *[]byte {
	var index int
	if size%p.sliceSize == 0 {
		index = size / p.sliceSize
	} else {
		index = size/p.sliceSize + 1
	}
	if index > p.typeNum {
		return nil
	}
	return p.pools[index-1].Get().(*[]byte)
}

func (p *BytePool) Put(bytes *[]byte) {
	var index int
	length := len(*bytes)
	if length%p.sliceSize == 0 {
		index = length / p.sliceSize
	} else {
		index = length/p.sliceSize + 1
	}
	if index < p.typeNum {
		p.pools[index].Put(bytes)
	}
}
