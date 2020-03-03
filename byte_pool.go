package zhanio

import "sync"

var BytePool sync.Pool

func init() {
	BytePool.New = func() interface{} {
		return []byte{}
	}
}
