package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestHandler_Closed(t *testing.T) {
	lock := sync.Mutex{}
	lock.Lock()
	fmt.Println("134")
	lock.Unlock()
}
