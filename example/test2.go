package main

import "sync"

type BufferPool struct {
	sync.Pool
}

func init() {

}
